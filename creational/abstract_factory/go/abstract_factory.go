package main

import "fmt"

type OrderDao interface {
	SaveOrder()
}

type OrderDetailDao interface {
	SaveOrderDetail()
}

type DaoFactory interface {
	CreateOrderDao() OrderDao
	CreateOrderDetailDao() OrderDetailDao
}

type RDBOrderDao struct{}

func (dao *RDBOrderDao) SaveOrder() {
	fmt.Println("rdb save order")
}

type RDBOrderDetailDao struct{}

func (dao *RDBOrderDetailDao) SaveOrderDetail() {
	fmt.Println("rdb save order detail")
}

type RDBDaoFactory struct{}

func (RDBDaoFactory) CreateOrderDao() OrderDao {
	return &RDBOrderDao{}
}

func (RDBDaoFactory) CreateOrderDetailDao() OrderDetailDao {
	return &RDBOrderDetailDao{}
}

type RedisOrderDao struct{}

func (dao *RedisOrderDao) SaveOrder() {
	fmt.Println("redis save order")
}

type RedisOrderDetailDao struct{}

func (dao *RedisOrderDetailDao) SaveOrderDetail() {
	fmt.Println("redis save order detail")
}

type RedisDaoFactory struct{}

func (RedisDaoFactory) CreateOrderDao() OrderDao {
	return &RedisOrderDao{}
}

func (RedisDaoFactory) CreateOrderDetailDao() OrderDetailDao {
	return &RedisOrderDetailDao{}
}

func main() {
	var factory DaoFactory
	factory = RDBDaoFactory{}
	factory.CreateOrderDao().SaveOrder()
	factory.CreateOrderDetailDao().SaveOrderDetail()

	factory = RedisDaoFactory{}
	factory.CreateOrderDao().SaveOrder()
	factory.CreateOrderDetailDao().SaveOrderDetail()
}
