from common_fixtures import *  # NOQA

import subprocess32
from subprocess32 import Popen
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

    def check_call(self, input, *args):
        print args
        p = self.call(*args)
        p.communicate(input=input, timeout=120)
        retcode = p.wait()
        assert 0 == retcode
        return p

    def call(self, *args):
        env = {
            'RANCHER_ACCESS_KEY': self.client._access_key,
            'RANCHER_SECRET_KEY': self.client._secret_key,
            'RANCHER_URL': self.client._url,
        }
        cmd = [self.compose_bin]
        cmd.extend(args)
        return Popen(cmd, env=env, stdin=subprocess32.PIPE, stdout=sys.stdout,
                     stderr=sys.stderr, cwd=_base())


@pytest.fixture
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
    for e in client.list_environment():
        client.delete(e)


@pytest.fixture(scope='session')
def compose(client, compose_bin, request):
    request.addfinalizer(lambda: _clean_all(client))
    return Compose(client, compose_bin)


def create_project(compose, file=None, input=None):
    project_name = random_str()
    if file is not None:
        compose.check_call(None, '-f', file, '-p', project_name, 'create')
    elif input is not None:
        compose.check_call(input, '-f', '-', '-p', project_name, 'create')

    PROJECTS.append(project_name)
    return project_name


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
    assert set(service.launchConfig.ports) == {'80:81/tcp', '123/tcp', '21/tcp'}
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
    assert service.launchConfig.name == 'foo'
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
    assert service.launchConfig.build == {
        'dockerfile': 'something/other',
        'remote': 'github.com/ibuildthecloud/tiny-build',
    }


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


def test_lb(client, compose):
    template = '''
lb:
    image: rancher/load-balancer
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

    assert lb.type == 'loadBalancerService'


def test_lb_full_config(client, compose):
    project_name = create_project(compose, file='assets/lb/docker-compose.yml')
    project = find_one(client.list_environment, name=project_name)
    assert len(project.services()) == 2

    lb = _get_service(project.services(), 'lb')
    web = _get_service(project.services(), 'web')

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
    image: nginx
db:
    image: mysql
    volumes_from:
    - web
'''
    project_name = create_project(compose, input=template)

    project = find_one(client.list_environment, name=project_name)

    web = _get_service(project.services(), 'web')
    db = _get_service(project.services(), 'db')

    assert web.dataVolumesFromService is None
    assert len(db.dataVolumesFromService) == 1
    assert db.dataVolumesFromService[0] == web.id


def _get_service(services, name):
    service = None

    for i in services:
        if i.name == name:
            service = i
            break

    assert service is not None
    return service


