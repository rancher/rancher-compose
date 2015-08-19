package upgrade

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	rancherClient "github.com/rancherio/go-rancher/client"
	"github.com/rancherio/rancher-compose/rancher"
)

type UpgradeOpts struct {
	BatchSize      int
	IntervalMillis int
	FinalScale     int
	UpdateLinks    bool
	Wait           bool
	Clean	       bool
}

func Upgrade(p *project.Project, from, to string, opts UpgradeOpts) error {
	fromService, err := p.CreateService(from)
	if err != nil {
		return err
	}

	toService, err := p.CreateService(to)
	if err != nil {
		return err
	}

	rFromService, ok := fromService.(*rancher.RancherService)
	if !ok {
		return fmt.Errorf("%s is not a Rancher service", from)
	}

	rToService, ok := toService.(*rancher.RancherService)
	if !ok {
		return fmt.Errorf("%s is not a Rancher service", to)
	}

	if service, err := rToService.RancherService(); err != nil {
		return err
	} else if service == nil {
		if err := rToService.Create(); err != nil {
			return err
		}

		if err := rToService.Scale(0); err != nil {
			return err
		}
	}

	if err := rToService.Up(); err != nil {
		return err
	}

	source, err := rFromService.RancherService()
	if err != nil {
		return err
	}

	dest, err := rToService.RancherService()
	if err != nil {
		return err
	}

	if source == nil {
		return fmt.Errorf("Failed to find service %s", from)
	}

	if dest == nil {
		return fmt.Errorf("Failed to find service %s", to)
	}

	upgradeOpts := &rancherClient.ServiceUpgrade{
		UpdateLinks:    opts.UpdateLinks,
		FinalScale:     int64(opts.FinalScale),
		BatchSize:      int64(opts.BatchSize),
		IntervalMillis: int64(opts.IntervalMillis),
		ToServiceId:    dest.Id,
	}
	if upgradeOpts.FinalScale == -1 {
		upgradeOpts.FinalScale = source.Scale
	}

	client := rFromService.Client()

	logrus.Infof("Upgrading %s to %s, scale=%d", from, to, upgradeOpts.FinalScale)
	service, err := client.Service.ActionUpgrade(source, upgradeOpts)
	if err != nil {
		return err
	}

	if opts.Wait {
		err := rFromService.Wait(service)
		if err != nil {
			return err
		}
	}

	if opts.Clean {
		err := rFromService.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func upgradeInfo(up bool, p *project.Project, from, to string, opts UpgradeOpts) (*rancherClient.Service, *rancherClient.Service, *rancherClient.RancherClient, error) {
	fromService, err := p.CreateService(from)
	if err != nil {
		return nil, nil, nil, err
	}

	toService, err := p.CreateService(to)
	if err != nil {
		return nil, nil, nil, err
	}

	rFromService, ok := fromService.(*rancher.RancherService)
	if !ok {
		return nil, nil, nil, fmt.Errorf("%s is not a Rancher service", from)
	}

	rToService, ok := toService.(*rancher.RancherService)
	if !ok {
		return nil, nil, nil, fmt.Errorf("%s is not a Rancher service", to)
	}

	if up {
		if err := rToService.Up(); err != nil {
			return nil, nil, nil, err
		}
	}

	source, err := rFromService.RancherService()
	if err != nil {
		return nil, nil, nil, err
	}

	dest, err := rToService.RancherService()
	if err != nil {
		return nil, nil, nil, err
	}

	return source, dest, rFromService.Client(), nil
}
