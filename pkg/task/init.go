/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心检索引擎
 * (BlueKing-IAM-Search-Engine) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package task

import (
	"sync"

	"engine/pkg/metric"
	"engine/pkg/redis"

	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"
)

const (
	ConnTypeProducer = "producer"
	ConnTypeConsumer = "consumer"
	ConnTypeCleaner  = "cleaner"
)

var (
	connection               rmq.Connection
	engineDeletionEventQueue rmq.Queue
)

var (
	connectionInitOnce               sync.Once
	engineDeletionEventQueueInitOnce sync.Once
)

const (
	rmqConnectionTag = "engine_rmq"
	rmqConsumerTag   = "consumer"
)

// InitRmqQueue 初始化rmq队列
func InitRmqQueue(debugMode bool) {
	errChan := make(chan error, 10)
	go logRmqErrors(errChan)

	var err error
	if connection == nil {
		connectionInitOnce.Do(func() {
			connection, err = rmq.OpenConnectionWithRedisClient(
				rmqConnectionTag,
				redis.GetDefaultMQRedisClient(),
				errChan,
			)
			if err != nil {
				log.WithError(err).Error("new rmq connection fail")
				if !debugMode {
					panic(err)
				}
			}

			// register metrics
			metric.RecordRmqMetrics(connection)
		})
	}

	if engineDeletionEventQueue == nil {
		engineDeletionEventQueueInitOnce.Do(func() {
			engineDeletionEventQueue, err = connection.OpenQueue("engine_deletion")
			if err != nil {
				log.WithError(err).Error("new rmq queue fail")
				if !debugMode {
					panic(err)
				}
			}
		})
	}
}

func logRmqErrors(errChan <-chan error) {
	for err := range errChan {
		log.WithError(err).Error("rmq error")
	}
}
