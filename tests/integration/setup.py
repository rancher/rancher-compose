from distutils.core import setup

setup(
    name='RancherCliIntegrationTests',
    version='0.1',
    packages=[
      'cattletest',
      'cattletest.core',
    ],
    license='ASL 2.0',
    long_description=open('README.txt').read(),
)
