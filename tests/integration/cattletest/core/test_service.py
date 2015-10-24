from common_fixtures import *  # NOQA

import subprocess
from subprocess import Popen
from os import path
import os

import sys
import pytest
import cattle
import ConfigParser


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


def test_stack_create_upgrade_finish(client):
    name = 'project-' + random_str()
    rancher_compose = '''
.catalog:
  uuid: foo
'''
    template = '''
one:
  image: nginx

two:
  image: nginx
'''

    env = client.create_environment(name=name, dockerCompose=template,
                                    rancherCompose=rancher_compose)
    env = client.wait_success(env)
    assert env.state == 'active'
    assert env.externalId == 'foo'

    names = set()
    for s in env.services():
        assert s.state == 'inactive'
        names.add(s.name)

    assert names == {'one', 'two'}

    env = client.wait_success(env.activateservices())
    for s in env.services():
        s = client.wait_success(s)
        assert s.state == 'active'

    rancher_compose = '''
.catalog:
  uuid: foo2
'''
    template = '''
one:
  image: nginx:2

two:
  image: nginx
'''

    # TODO: externalId should not be in upgrade
    env.upgrade(dockerCompose=template,
                rancherCompose=rancher_compose,
                externalId='foo2')

    env = client.wait_success(env)
    for s in env.services():
        s = client.wait_success(s)
        if s.name == 'one':
            assert s.state == 'upgraded'
        elif s.name == 'two':
            assert s.state == 'active'

    assert env.externalId == 'foo2'
    assert env.previousExternalId == 'foo'

    env.finishupgrade()

    env = client.wait_success(env)
    for s in env.services():
        s = client.wait_success(s)
        assert s.state == 'active'

    assert env.externalId == 'foo2'
    assert env.previousExternalId is None


def test_stack_create_and_upgrade(client):
    name = 'project-' + random_str()
    rancher_compose = '''
.catalog:
  uuid: foo
'''
    template = '''
one:
  image: nginx

two:
  image: nginx
'''

    env = client.create_environment(name=name, dockerCompose=template,
                                    rancherCompose=rancher_compose)
    env = client.wait_success(env)
    env = client.wait_success(env.activateservices())
    assert env.state == 'active'
    for s in env.services():
        s = client.wait_success(s)
        assert s.state == 'active'

    rancher_compose = '''
.catalog:
  uuid: foo2
'''
    template = '''
one:
  image: nginx:2

two:
  image: nginx
'''

    # TODO: externalId should not be in upgrade
    env.upgrade(dockerCompose=template,
                rancherCompose=rancher_compose,
                externalId='foo2')

    env = client.wait_success(env)
    for s in env.services():
        s = client.wait_success(s)
        if s.name == 'one':
            assert s.state == 'upgraded'

    env.rollback()
    env = client.wait_success(env)
    for s in env.services():
        s = client.wait_success(s)
        assert s.state == 'active'

    assert env.externalId == 'foo'
    assert env.previousExternalId is None


def _base():
    return path.dirname(__file__)

