package main

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

type RDBOrderDao struct {
}

func (dao *RDBOrderDao) SaveOrder() {
	println("rdb save order")
}

type RDBOrderDetailDao struct {
}

func (dao *RDBOrderDetailDao) SaveOrderDetail() {
	println("rdb save order detail")
}

type RDBDaoFactory struct {
}

func (RDBDaoFactory) CreateOrderDao() OrderDao {
	return &RDBOrderDao{}
}

func (RDBDaoFactory) CreateOrderDetailDao() OrderDetailDao {
	return &RDBOrderDetailDao{}
}

type RedisDaoFactory struct {
}

func (RedisDaoFactory) CreateOrderDao() OrderDao {
	return &RedisOrderDao{}
}

func (RedisDaoFactory) CreateOrderDetailDao() OrderDetailDao {
	return &RedisOrderDetailDao{}
}

type RedisOrderDao struct {
}

func (dao *RedisOrderDao) SaveOrder() {
	println("redis save order")
}

type RedisOrderDetailDao struct {
}

func (dao *RedisOrderDetailDao) SaveOrderDetail() {
	println("redis save order detail")
}

func main() {
	var factory DaoFactory
	factory = RDBDaoFactory{}
	orderDao := factory.CreateOrderDao()
	orderDetailDao := factory.CreateOrderDetailDao()
	orderDao.SaveOrder()
	orderDetailDao.SaveOrderDetail()

	factory = RedisDaoFactory{}
	orderDao = factory.CreateOrderDao()
	orderDetailDao = factory.CreateOrderDetailDao()
	orderDao.SaveOrder()
	orderDetailDao.SaveOrderDetail()
}
