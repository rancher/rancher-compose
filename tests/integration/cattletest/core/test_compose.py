from common_fixtures import *  # NOQA

import subprocess
from subprocess import Popen
from os import path
import os

import sys
import pytest
import cattle
import ConfigParser


PROJECTS = []


class Compose(object):
    def __init__(self, client, compose_bin):
        self.compose_bin = compose_bin
        self.client = client

    def check_retcode(self, input, check_retcode, *args):
        p = self.call(*args)
        p.communicate(input=input)
        retcode = p.wait()
        assert check_retcode == retcode
        return p

    def check_call(self, input, *args):
        p = self.call(*args)
        p.communicate(input=input)
        retcode = p.wait()
        assert 0 == retcode
        return p

    def call(self, *args):
        env = {
            'RANCHER_CLIENT_DEBUG': 'true',
            'RANCHER_ACCESS_KEY': self.client._access_key,
            'RANCHER_SECRET_KEY': self.client._secret_key,
            'RANCHER_URL': self.client._url,
        }
        cmd = [self.compose_bin]
        cmd.extend(args)
        return Popen(cmd, env=env, stdin=subprocess.PIPE, stdout=sys.stdout,
                     stderr=sys.stderr, cwd=_base())


@pytest.fixture(scope='session')
def client(admin_user_client, request):
    try:
        return cattle.from_env(url=os.environ['RANCHER_URL'],
                               access_key=os.environ['RANCHER_ACCESS_KEY'],
                               secret_key=os.environ['RANCHER_SECRET_KEY'])
    except KeyError:
        pass

    try:
        config = ConfigParser.ConfigParser()
        config.read(path.join(_base(), '../../tox.ini'))
        return cattle.from_env(url=config.get('rancher', 'url'),
                               access_key=config.get('rancher', 'access-key'),
                               secret_key=config.get('rancher', 'secret-key'))
    except ConfigParser.NoOptionError:
        pass

    return new_context(admin_user_client, request).client


def _file(f):
    return path.join(_base(), '../../../../{}'.format(f))


def _base():
    return path.dirname(__file__)


@pytest.fixture(scope='session')
def compose_bin():
    c = _file('bin/rancher-compose')
    assert path.exists(c)
    return c


def _clean_all(client):
    for p in PROJECTS:
        client.delete(p)


@pytest.fixture(scope='session')
def compose(client, compose_bin, request):
    request.addfinalizer(lambda: _clean_all(client))
    return Compose(client, compose_bin)


def create_project(compose, operation='create', project_name=None, file=None,
                   input=None):
    if project_name is None:
        project_name = random_str()
    if file is not None:
        compose.check_call(None, '--verbose', '-f', file, '-p', project_name,
                           operation)
    elif input is not None:
        compose.check_call(input, '--verbose', '-f', '-', '-p', project_name,
                           operation)

    PROJECTS.append(project_name)
    return project_name


@pytest.mark.skipif('True')
def test_build(client, compose):
    project_name = create_project(compose, file='assets/build/test.yml')

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.name == 'fromfile'
    assert service.launchConfig.build.dockerfile == 'subdir/Dockerfile'
    assert service.launchConfig.build.remote is None
    assert service.launchConfig.build.context.startswith('https://')


def test_args(client, compose):
    project_name = create_project(compose, file='assets/full.yml')
    project = find_one(client.list_environment, name=project_name)
    assert project.name == project_name

    service = find_one(project.services)
    assert service.name == 'web'
    assert service.launchConfig.command == ['/bin/sh', '-c']
    assert service.launchConfig.imageUuid == 'docker:nginx'
    assert set(service.launchConfig.ports) == {'80:81/tcp', '123/tcp',
                                               '21/tcp'}
    assert service.launchConfig.dataVolumes == ['/tmp/foo', '/tmp/x:/tmp/y']
    assert service.launchConfig.environment == {'foo': 'bar', 'a': 'b'}
    assert service.launchConfig.dns == ['8.8.8.8', '1.1.1.1']
    assert service.launchConfig.capAdd == ['ALL', 'SYS_ADMIN']
    assert service.launchConfig.capDrop == ['NET_ADMIN', 'SYS_ADMIN']
    assert service.launchConfig.dnsSearch == ['foo.com', 'bar.com']
    assert service.launchConfig.entryPoint == ['/bin/foo', 'bar']
    assert service.launchConfig.workingDir == '/somewhere'
    assert service.launchConfig.user == 'somebody'
    assert service.launchConfig.hostname == 'myhostname'
    assert service.launchConfig.domainName == 'example.com'
    assert service.launchConfig.memory == 100
    assert service.launchConfig.memorySwap == 101
    assert service.launchConfig.privileged
    assert service.launchConfig.restartPolicy == {
        'name': 'always'
    }
    assert service.launchConfig.stdinOpen
    assert service.launchConfig.tty
    assert 'name' not in service.launchConfig
    assert service.launchConfig.cpuShares == 42
    assert service.launchConfig.cpuSet == '1,2'
    assert service.launchConfig.devices == ['/dev/sda:/dev/a:rwm',
                                            '/dev/sdb:/dev/c:ro']
    assert service.launchConfig.labels == {'a': 'b', 'c': 'd'}
    assert service.launchConfig.securityOpt == ['label:foo', 'label:bar']
    assert service.launchConfig.pidMode == 'host'
    assert service.launchConfig.logConfig == {
        'driver': 'syslog',
        'config': {
            'tag': 'foo',
        }
    }
    assert service.launchConfig.extraHosts == ['host:1.1.1.1', 'host:2.2.2.2']
    assert service.launchConfig.networkMode == 'host'
    assert service.launchConfig.volumeDriver == 'foo'
    assert service.launchConfig.build == {
        'dockerfile': 'something/other',
        'remote': 'github.com/ibuildthecloud/tiny-build',
    }
    assert service.launchConfig.externalLinks == ['foo', 'bar']


def test_git_build(client, compose):
    template = '''
    nginx:
      build: github.com/ibuildthecloud/tiny-build
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.build == {
        'remote': 'github.com/ibuildthecloud/tiny-build',
    }
    assert service.launchConfig.imageUuid is not None

    prefix = 'docker:{}_nginx_'.format(project_name)
    assert service.launchConfig.imageUuid.startswith(prefix)


def test_circular_sidekick(client, compose):
    template = '''
    primary:
      stdin_open: true
      image: busybox
      command: cat
      labels:
        io.rancher.sidekicks: secondary
      volumes_from:
      - secondary

    secondary:
      stdin_open: true
      image: busybox
      command: cat
    '''
    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.dataVolumesFromLaunchConfigs == ['secondary']
    secondary = filter(lambda x: x.name == 'secondary',
                       service.secondaryLaunchConfigs)
    assert len(secondary) == 1


def test_delete(client, compose):
    template = '''
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.state == 'inactive'

    compose.check_call(template, '--verbose', '-f', '-', '-p', project_name,
                       'up', '-d')

    service = client.wait_success(service)

    assert service.state == 'active'

    compose.check_call(template, '--verbose', '-f', '-', '-p', project_name,
                       'rm', '--force')

    service = client.wait_success(service)

    assert service.state == 'removed'


def test_delete_while_stopped(client, compose):
    template = '''
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.state == 'inactive'

    compose.check_call(template, '--verbose', '-f', '-', '-p', project_name,
                       'rm', 'web')

    service = client.wait_success(service)

    assert service.state == 'removed'


def test_network_bridge(client, compose):
    template = '''
    web:
        net: bridge
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.networkMode == 'bridge'


def test_network_none(client, compose):
    template = '''
    web:
        net: none
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.networkMode == 'none'


def test_network_container(compose, client):
    template = '''
    foo:
        labels:
            io.rancher.sidekicks: web
        image: nginx

    web:
        net: container:foo
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.networkMode == 'managed'
    assert service.secondaryLaunchConfigs[0].networkMode == 'container'
    assert service.secondaryLaunchConfigs[0].networkLaunchConfig == 'foo'


def test_network_managed(client, compose):
    template = '''
    web:
        net: managed
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.networkMode == 'managed'


def test_network_default(client, compose):
    template = '''
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.networkMode == 'managed'


def test_env_file(client, compose):
    project_name = create_project(compose, file='assets/base.yml')
    project = find_one(client.list_environment, name=project_name)
    assert project.name == project_name

    second = _get_service(project.services(), 'base')

    assert second.launchConfig.environment == {
        'bar': 'baz',
        'd': 'e',
        'env': '2',
        'foo': 'bar',
        'a': 'b',
    }


def test_extends(client, compose):
    project_name = create_project(compose, file='assets/base.yml')
    project = find_one(client.list_environment, name=project_name)
    assert project.name == project_name

    base = _get_service(project.services(), 'base')
    local = _get_service(project.services(), 'local')
    other_base = _get_service(project.services(), 'other-base')

    assert base.launchConfig.imageUuid == 'docker:second'

    assert local.launchConfig.imageUuid == 'docker:local'
    assert local.launchConfig.ports == ['80/tcp']
    assert local.launchConfig.environment == {'key': 'value'}

    assert other_base.launchConfig.ports == ['80/tcp', '81/tcp']
    assert other_base.launchConfig.imageUuid == 'docker:other'
    assert other_base.launchConfig.environment == {'key': 'value',
                                                   'key2': 'value2'}


def test_extends_1556(client, compose):
    project_name = create_project(compose,
                                  file='assets/extends/docker-compose.yml')
    project = find_one(client.list_environment, name=project_name)
    assert project.name == project_name

    web = _get_service(project.services(), 'web')
    db = _get_service(project.services(), 'db')

    # Notice the images are wrong
    assert web.launchConfig.imageUuid == 'docker:ubuntu:14:04'
    assert db.launchConfig.imageUuid == 'docker:ubuntu:14:04'

    web = find_one(db.consumedservices)
    assert web.name == 'web'


def test_extends_1556_2(compose):
    with pytest.raises(AssertionError):
        create_project(compose, file='assets/extends_2/docker-compose.yml')


def test_restart_policies(client, compose):
    template = '''
    web:
        restart: on-failure:5
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.restartPolicy == {
        'name': 'on-failure',
        'maximumRetryCount': 5
    }


def test_restart_policies_on_failure_default(client, compose):
    template = '''
    web:
        restart: on-failure
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.launchConfig.restartPolicy == {
        'name': 'on-failure'
    }


def test_lb_basic(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 80
        links:
        - web
        - web2
    web:
        image: nginx
    web2:
        image: nginx'''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 3
    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')
    web2 = _get_service(project.services(), 'web2')

    maps = client.list_service_consume_map(serviceId=lb.id)
    assert len(maps) == 2

    for map in maps:
        if map.consumedServiceId == web.id:
            assert map.ports == []
        elif map.consumedServiceId == web2.id:
            assert map.ports == []
        else:
            assert False

    assert lb.type == 'loadBalancerService'
    assert lb.launchConfig.ports == ['80']


def test_lb_default_port_http(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 7900:80/tcp
        links:
        - web
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2
    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')
    assert lb.launchConfig.ports == ['7900:80/tcp']

    map = find_one(client.list_service_consume_map, serviceId=lb.id)
    assert map.consumedServiceId == web.id
    assert map.ports == []

    assert lb.launchConfig.ports == ['7900:80/tcp']


def test_lb_default_port_with_mapped_tcp(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 80:8080/tcp
        links:
        - web
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2
    lb = _get_service(project.services(), 'lb')
    assert lb.launchConfig.ports == ['80:8080/tcp']

    web = _get_service(project.services(), 'web')

    map = find_one(client.list_service_consume_map, serviceId=lb.id)
    assert map.consumedServiceId == web.id
    assert map.ports == []

    assert lb.launchConfig.ports == ['80:8080/tcp']


def test_lb_default_port_with_tcp(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 80/tcp
        links:
        - web
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2
    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')

    map = find_one(client.list_service_consume_map, serviceId=lb.id)
    assert map.consumedServiceId == web.id
    assert map.ports == []

    lb.launchConfig.ports == ['80/tcp']


def test_lb_path_space_target(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 80:8080
        labels:
          io.rancher.loadbalancer.target.web: "hostname/path:6000,
           7000"
        links:
        - web
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2
    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')

    maps = client.list_service_consume_map(serviceId=lb.id)
    assert len(maps) == 1

    for map in maps:
        if map.consumedServiceId == web.id:
            assert map.ports == ['hostname/path:6000',
                                 '7000']
        else:
            assert False

    assert lb.type == 'loadBalancerService'


def test_lb_path_name(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 80:8080
        labels:
          io.rancher.loadbalancer.target.web: hostname/path:6000,hostname:7000
          io.rancher.loadbalancer.target.web2: 9000
        links:
        - web
        - web2
        - web3
    web:
        image: nginx
    web2:
        image: nginx
    web3:
        image: nginx'''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 4
    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')
    web2 = _get_service(project.services(), 'web2')
    web3 = _get_service(project.services(), 'web2')

    maps = client.list_service_consume_map(serviceId=lb.id)
    assert len(maps) == 3

    for map in maps:
        if map.consumedServiceId == web.id:
            assert map.ports == ['hostname/path:6000',
                                 'hostname:7000']
        elif map.consumedServiceId == web2.id:
            assert map.ports == ['9000']
        elif map.consumedServiceId == web3.id:
            assert map.ports == []

    assert lb.launchConfig.ports == ['80:8080']
    assert lb.type == 'loadBalancerService'


def test_lb_path_name_minimal(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 84
        links:
        - web
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2
    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')

    map = find_one(client.list_service_consume_map, serviceId=lb.id)
    assert map.ports == []
    assert map.consumedServiceId == web.id

    assert lb.type == 'loadBalancerService'
    assert lb.launchConfig.ports == ['84']


def test_lb_full_config(client, compose):
    project_name = create_project(compose, file='assets/lb/docker-compose.yml')
    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2

    lb = _get_service(project.services(), 'lb')
    _get_service(project.services(), 'web')

    assert lb.type == 'loadBalancerService'

    assert lb.loadBalancerConfig.name == 'lb config'
    assert lb.loadBalancerConfig.appCookieStickinessPolicy.cookie == 'foo'
    assert lb.loadBalancerConfig.appCookieStickinessPolicy.maxLength == 1024
    assert 'prefix' not in lb.loadBalancerConfig.appCookieStickinessPolicy
    assert lb.loadBalancerConfig.appCookieStickinessPolicy.requestLearn
    assert lb.loadBalancerConfig.appCookieStickinessPolicy.mode == \
        'path_parameters'
    assert 'port' not in lb.loadBalancerConfig.healthCheck
    assert lb.loadBalancerConfig.healthCheck.interval == 2000
    assert lb.loadBalancerConfig.healthCheck.unhealthyThreshold == 3
    assert lb.loadBalancerConfig.healthCheck.requestLine == \
        'OPTIONS /ping HTTP/1.1\r\nHost:\\ www.example.com'
    assert lb.loadBalancerConfig.healthCheck.healthyThreshold == 2
    assert lb.loadBalancerConfig.healthCheck.responseTimeout == 2000


def test_links(client, compose):
    template = '''
    web:
        image: nginx
    db:
        image: mysql
        links:
        - web
    other:
        image: foo
        links:
        - web
        - db
    '''

    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)

    web = _get_service(project.services(), 'web')
    db = _get_service(project.services(), 'db')
    other = _get_service(project.services(), 'other')

    assert len(web.consumedservices()) == 0

    db_consumed = db.consumedservices()
    assert len(db_consumed) == 1
    assert db_consumed[0].name == 'web'

    other_consumed = other.consumedservices()
    assert len(other_consumed) == 2
    names = {i.name for i in other_consumed}
    assert names == {'web', 'db'}


def test_volumes_from(client, compose):
    template = '''
    web:
        labels:
            io.rancher.sidekicks: db
        image: nginx
    db:
        image: mysql
        volumes_from:
        - web
    '''
    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.secondaryLaunchConfigs[0].dataVolumesFromLaunchConfigs == \
        ['web']


def test_sidekick_simple(client, compose):
    template = '''
    web:
        labels:
            io.rancher.sidekicks: log
        image: nginx
    log:
        image: mysql
    log2:
        image: bar
    '''
    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    services = project.services()

    service = _get_service(services, 'web')
    log2 = _get_service(services, 'log2')

    assert len(services) == 2
    assert service.name == 'web'
    assert service.launchConfig.imageUuid == 'docker:nginx'
    assert service.launchConfig.networkMode == 'managed'
    assert len(service.secondaryLaunchConfigs) == 1
    assert service.secondaryLaunchConfigs[0].name == 'log'
    assert service.secondaryLaunchConfigs[0].imageUuid == 'docker:mysql'
    assert service.secondaryLaunchConfigs[0].networkMode == 'managed'

    assert log2.name == 'log2'
    assert log2.launchConfig.imageUuid == 'docker:bar'


def test_sidekick_container_network(client, compose):
    template = '''
    web:
        labels:
            io.rancher.sidekicks: log
        image: nginx
    log:
        net: container:web
        image: mysql
    '''
    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.name == 'web'
    assert service.launchConfig.imageUuid == 'docker:nginx'
    assert len(service.secondaryLaunchConfigs) == 1
    assert service.secondaryLaunchConfigs[0].name == 'log'
    assert service.secondaryLaunchConfigs[0].imageUuid == 'docker:mysql'
    assert service.secondaryLaunchConfigs[0].networkMode == 'container'
    assert service.secondaryLaunchConfigs[0].networkLaunchConfig == 'web'


def test_external_service_hostname(client, compose):
    project_name = create_project(compose, file='assets/hostname/test.yml')

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.name == 'web'
    assert service.type == 'externalService'
    assert 'launchConfig' not in service
    assert service.hostname == 'example.com'


def test_external_ip(client, compose):
    project_name = create_project(compose, file='assets/externalip/test.yml')

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.name == 'web'
    assert service.type == 'externalService'
    assert 'launchConfig' not in service
    assert service.externalIpAddresses == ['1.1.1.1', '2.2.2.2']


def test_dns_service(client, compose):
    template = '''
    web1:
        image: nginx
    web2:
        image: nginx
    web:
        image: rancher/dns-service
        links:
        - web1
        - web2
    '''
    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)
    services = project.services()

    assert len(services) == 3

    web = _get_service(services, 'web')

    assert web.type == 'dnsService'
    consumed = web.consumedservices()

    assert len(consumed) == 2
    names = {x.name for x in consumed}

    assert names == {'web1', 'web2'}


def test_up_relink(client, compose):
    template = '''
    lb:
        image: rancher/load-balancer-service
        ports:
        - 80
        links:
        - web
        labels:
          a: b
          c: d
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    lb = _get_service(project.services(), 'lb')

    consumed = lb.consumedservices()
    assert len(consumed) == 1
    assert consumed[0].name == 'web'

    assert lb.launchConfig.labels == {
        'a': 'b',
        'c': 'd',
    }

    template2 = '''
    lb:
        image: nginx
        ports:
        - 80
        links:
        - web2
    web2:
        image: nginx
    '''
    compose.check_call(template2, '--verbose', '-f', '-', '-p', project_name,
                       'up', '-d')

    def check():
        x = lb.consumedservices()
        if len(x) == 1:
            return x

    consumed = wait_for(check, timeout=5)
    assert len(consumed) == 1
    assert consumed[0].name == 'web2'


def test_service_upgrade_from_nil(client, compose):
    template = '''
    foo:
        image: nginx
    web2:
        image: nginx
    '''

    project_name = create_project(compose, input=template)

    upgrade = '''
    foo:
        image: nginx
    web:
        image: nginx
    web2:
        image: nginx
    '''

    compose.check_retcode(upgrade, 1, '-p', project_name, '-f',
                          '-', 'upgrade', 'web', 'web2')


def test_service_map_syntax(client, compose):
    template = '''
    foo:
        image: nginx
        links:
            web: alias
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    foo = _get_service(project.services(), 'foo')
    maps = client.list_serviceConsumeMap(serviceId=foo.id)

    assert len(maps) == 1
    assert maps[0].name == 'alias'


def test_cross_stack_link(client, compose):
    template = '''
    dest:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    dest = _get_service(project.services(), 'dest')

    template = '''
    src:
        external_links:
        - {}/dest
        image: nginx
    '''.format(project_name)

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    src = _get_service(project.services(), 'src')

    services = src.consumedservices()
    assert len(services) == 1

    assert services[0].id == dest.id


def test_upgrade_ignore_scale(client, compose):
    project_name = create_project(compose, file='assets/upgrade-ignore-scale/'
                                                'docker-compose-source.yml')
    compose.check_call(None, '--verbose', '-f', 'assets/upgrade-ignore-scale/'
                       'docker-compose-source.yml',
                       '-p', project_name, 'up', '-d')
    project = find_one(client.list_environment, name=project_name)
    compose.check_call(None, '-p', project_name, '-f',
                       'assets/upgrade-ignore-scale/docker-compose.yml',
                       'upgrade', '--scale=2', 'from', 'to')

    f = _get_service(project.services(), 'from')
    to = _get_service(project.services(), 'to')

    assert to.scale == 0

    f = client.wait_success(f)
    to = client.wait_success(to)

    assert f.scale == 0
    assert to.scale == 2


def test_service_link_with_space(client, compose):
    template = '''
    foo:
        image: nginx
        links:
        - "web: alias"
    web:
        image: nginx
    '''

    project_name = create_project(compose, input=template)
    project = find_one(client.list_environment, name=project_name)
    foo = _get_service(project.services(), 'foo')
    maps = client.list_serviceConsumeMap(serviceId=foo.id)

    assert len(maps) == 1
    assert maps[0].name == 'alias'


def test_circle(client, compose):
    template = '''
    foo:
        image: nginx
        links:
        - web
    web:
        image: nginx
        links:
        - foo
    '''

    project_name = random_str()
    compose.check_call(template, '-p', project_name, '-f',
                       '-', 'up', '-d')
    project = find_one(client.list_environment, name=project_name)
    foo = _get_service(project.services(), 'foo')
    web = _get_service(project.services(), 'web')

    s = find_one(foo.consumedservices)
    assert s.name == 'web'

    s = find_one(web.consumedservices)
    assert s.name == 'foo'


def test_one_circle(client, compose):
    template = '''
    foo:
        image: nginx
        links:
        - foo
    '''

    project_name = random_str()
    compose.check_call(template, '-p', project_name, '-f',
                       '-', 'up', '-d')
    project = find_one(client.list_environment, name=project_name)
    foo = _get_service(project.services(), 'foo')

    s = find_one(foo.consumedservices)
    assert s.name == 'foo'


def test_circle_madness(client, compose):
    template = '''
    foo:
        image: nginx
        links:
        - foo
        - foo2
        - foo3
    foo2:
        image: nginx
        links:
        - foo
        - foo2
        - foo3
    foo3:
        image: nginx
        links:
        - foo
        - foo2
        - foo3
    '''

    project_name = random_str()
    compose.check_call(template, '-p', project_name, '-f',
                       '-', 'up', '-d')
    project = find_one(client.list_environment, name=project_name)
    foo = _get_service(project.services(), 'foo')
    foo2 = _get_service(project.services(), 'foo2')
    foo3 = _get_service(project.services(), 'foo3')

    assert len(foo.consumedservices()) == 3
    assert len(foo2.consumedservices()) == 3
    assert len(foo3.consumedservices()) == 3


def test_healthchecks(client, compose):
    project_name = create_project(compose, file='assets/health/test.yml')

    project = find_one(client.list_environment, name=project_name)
    service = find_one(project.services)

    assert service.name == 'web'
    assert service.launchConfig.healthCheck.port == 80
    assert service.launchConfig.healthCheck.interval == 2000
    assert service.launchConfig.healthCheck.unhealthyThreshold == 3
    assert service.launchConfig.healthCheck.requestLine == \
        "OPTIONS /ping HTTP/1.1\r\nHost:\\ www.example.com"


def _get_service(services, name):
    service = None

    for i in services:
        if i.name == name:
            service = i
            break

    assert service is not None
    return service
