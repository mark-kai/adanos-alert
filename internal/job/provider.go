package job

import (
	"fmt"
	"github.com/mylxsw/adanos-alert/internal/repository"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"sync"

	"github.com/mylxsw/adanos-alert/configs"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/glacier/cron"
)

type ServiceProvider struct{}

func (s ServiceProvider) Register(app container.Container) {
	app.MustSingleton(NewAggregationJob)
	app.MustSingleton(NewTrigger)
}

func (s ServiceProvider) Boot(app glacier.Glacier) {
	app.Cron(func(cr cron.Manager, cc container.Container) error {

		return cc.Resolve(func(conf *configs.Config, aggregationJob *AggregationJob, alertJob *TriggerJob, lockRepo repository.LockRepo) {
			hostname, _ := os.Hostname()
			cr.DistributeLockManager(NewDistributeLockManager(lockRepo, fmt.Sprintf("%s(%s)", hostname, conf.Listen)))

			_ = cr.Add(AggregationJobName, fmt.Sprintf("@every %s", conf.AggregationPeriod), aggregationJob.Handle)
			_ = cr.Add(TriggerJobName, fmt.Sprintf("@every %s", conf.ActionTriggerPeriod), alertJob.Handle)
		})
	})
}

type DistributeLockManager struct {
	syncLock sync.RWMutex
	lockRepo repository.LockRepo
	lockID   primitive.ObjectID
	locked   bool
	owner    string
}

func NewDistributeLockManager(lockRepo repository.LockRepo, owner string) cron.DistributeLockManager {
	return &DistributeLockManager{lockRepo: lockRepo, locked: false, owner: owner}
}

var lockResource = "crontab-lock"

func (d *DistributeLockManager) TryLock() error {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	if d.locked {
		if _, err := d.lockRepo.Renew(d.lockID, 90); err != nil {
			if err == repository.ErrLockNotFound {
				if err := d.lock(); err != nil {
					return err
				}
			} else {
				return errors.Wrap(err, "renew lock failed")
			}
		}
	} else {
		if err := d.lock(); err != nil {
			return err
		}
	}

	return nil
}

func (d *DistributeLockManager) lock() error {
	lock, err := d.lockRepo.Lock(lockResource, d.owner, 90)
	if err != nil {
		if err == repository.ErrAlreadyLocked {
			d.lockID = primitive.NilObjectID
			d.locked = false
			return nil
		}

		return errors.Wrap(err, "acquire lock failed")
	}

	d.lockID = lock.LockID
	d.locked = true

	return nil
}

func (d *DistributeLockManager) TryUnLock() error {
	d.syncLock.Lock()
	defer d.syncLock.Unlock()

	if d.locked {
		if err := d.lockRepo.UnLock(d.lockID); err != nil {
			d.locked = false
			d.lockID = primitive.NilObjectID
			return err
		}
	}

	return nil
}

func (d *DistributeLockManager) HasLock() bool {
	d.syncLock.RLock()
	defer d.syncLock.RUnlock()

	return d.locked
}
