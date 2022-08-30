package main

import (
	"fmt"
	"github.com/maczh/mgconfig"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"time"
)

func main() {
	mgconfig.InitConfig("test.yml")
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//for i := 0; i < 100; i++ {
	//	go testMgoreplicaSet(i, r.Intn(300))
	//}
	//time.Sleep(time.Duration(5) * time.Minute)
	//var partnerInfos []PartnerInfo
	//mgconfig.GetMysqlConnection().Table("partner_info").Find(&partnerInfos)
	//redis,err := mgconfig.GetRedisConnection()
	//if err != nil {
	//	fmt.Println("redis连接失败:"+err.Error())
	//	mgconfig.SafeExit()
	//	return
	//}
	//for _, p := range partnerInfos {
	//	err =redis.Set(fmt.Sprintf("partner:info:%d",p.Id),utils.ToJSON(p),0).Err()
	//	if err != nil {
	//		fmt.Println("redis插入失败:"+err.Error())
	//	}
	//}

	//dbNames := []string{"flameCloud","cdnCloud","idc666"}
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//for i:= 1; i<=100; i++ {
	//	user := Users{
	//		Id:       i,
	//		Username: fmt.Sprintf("测试用户%d", i),
	//	}
	//	db,err := mgconfig.GetMySQLConnection(dbNames[r.Intn(3)])
	//	if err != nil {
	//		fmt.Printf("%s",err.Error())
	//		continue
	//	}
	//	db.Save(user)
	//}
	//for i:= 101; i<= 200; i++ {
	//	user := Users{
	//		Id: i,
	//		Username: fmt.Sprintf("user_%d",i),
	//	}
	//	db,err := mgconfig.GetMySQLConnection()
	//	if err != nil {
	//		fmt.Printf("%s",err.Error())
	//		continue
	//	}
	//	db.Save(user)
	//}
	//var rds *redis.Client
	//var err error
	//rds,err = mgconfig.GetRedisConnection("test2")
	//if err != nil {
	//	fmt.Printf("%s",err.Error())
	//}
	//for i := 0; i < 100; i++ {
	//	if i == 51 {
	//		rds,err = mgconfig.GetRedisConnection("test3")
	//		if err != nil {
	//			fmt.Printf("%s",err.Error())
	//		}
	//	}
	//	rds.SAdd("users:login",fmt.Sprintf("user_%d",i))
	//}
	mgconfig.RabbitCreateNewQueue("queue1")
	mgconfig.RabbitCreateDeadLetterQueue("queue1_dead","queue1",6000)
	for i := 0; i < 10; i++ {
		mgconfig.RabbitSendMessage("queue1_dead", fmt.Sprintf("Test Dead Letter %d",i))
		time.Sleep(time.Second)
	}
	mgconfig.SafeExit()
}

type Users struct {
	Id bson.ObjectId `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
}

func (Users) TableName() string{
	return "users"
}

type PartnerInfo struct {
	Id             int    `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT" bson:"id"`
	PartnerName    string `json:"partnerName" gorm:"column:partner_name" bson:"partnerName"`          // 代理商名称
	CompanyName    string `json:"companyName" gorm:"column:company_name" bson:"companyName"`          // 公司名称
	CompanyAddress string `json:"companyAddress" gorm:"column:company_address" bson:"companyAddress"` // 公司地址
	CreditCode     string `json:"creditCode" gorm:"column:credit_code" bson:"creditCode"`             // 企业信用代码证号
	ContactName    string `json:"contactName" gorm:"column:contact_name" bson:"contactName"`          // 企业联系人名称
	ParentId       int    `json:"parentId" gorm:"column:parent_id" bson:"parentId"`                   // 上级代理商编号
	WarningMobile  string `json:"warningMobile" gorm:"column:warning_mobile" bson:"warningMobile"`    // 告警手机号
	WarningEmail   string `json:"warningEmail" gorm:"column:warning_email" bson:"warningEmail"`       // 告警邮箱账号
	BillEmail      string `json:"billEmail" gorm:"column:bill_email" bson:"billEmail"`                // 账单通知邮箱账号
	Level          int    `json:"level" gorm:"column:level" bson:"level"`                             // 代理商级别 0-总代理 1-一级代理 2-二级代理
	Status         int    `json:"status" gorm:"column:status" bson:"status"`                          // 代理商状态 1-正常 2-欠费暂停 3-停止合作
	AdminMobile    string `json:"adminMobile" gorm:"column:admin_mobile" bson:"adminMobile"`          // 管理员手机号
	AdminEmail     string `json:"adminEmail" gorm:"column:admin_email" bson:"adminEmail"`             // 管理员邮箱账号
	AdminPwd       string `json:"-" gorm:"column:admin_pwd" bson:"adminPwd"`                          // 管理员登录密码
}

func (PartnerInfo) TableName() string {
	return "partner_info"
}

func testGormV2(thread, seconds int) {
	ids := []int{100007, 100008, 100009, 100010}
	var partnerInfo PartnerInfo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < seconds; i++ {
		partnerInfo.Id = ids[r.Intn(4)]
		mgconfig.GetMysqlConnection().First(&partnerInfo)
		fmt.Printf("线程:%d, ", thread)
		fmt.Println(partnerInfo)
		time.Sleep(time.Second)
	}
}

func testMgoreplicaSet(thread, seconds int) {
	ids := []int{100007, 100008, 100009, 100010}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < seconds; i++ {
		id := ids[r.Intn(4)]
		partnerInfo := getPartnerInfoMgo(thread,id)
		fmt.Printf("线程:%d, ", thread)
		fmt.Println(partnerInfo)
		time.Sleep(time.Second)
	}
}

func testRedis(thread, seconds int) {
	ids := []int{100007, 100008, 100009, 100010}
	//var partnerInfo PartnerInfo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < seconds; i++ {
		id := ids[r.Intn(4)]
		pi := getPartnerInfoRedis(thread, id)
		fmt.Printf("线程:%d, %s\n", thread, pi)
		time.Sleep(time.Second)
	}
}


func getPartnerInfoMgo(thread, id int) PartnerInfo {
	var partnerInfo PartnerInfo
	mongo, err := mgconfig.GetMongoConnection()
	if err != nil {
		fmt.Printf("线程:%d,MongoDB连接失败:%s ", thread, err.Error())
		return partnerInfo
	}
	mongo.C("PartnerInfo").Find(bson.M{"id": id}).One(&partnerInfo)
	mgconfig.ReturnMongoConnection(mongo)
	return partnerInfo
}

func getPartnerInfoRedis(thread, id int) string {
	redis, err := mgconfig.GetRedisConnection()
	if err != nil {
		fmt.Printf("线程:%d,Redis连接失败:%s", thread, err.Error())
		return ""
	}
	return redis.Get(fmt.Sprintf("partner:info:%d", id)).String()
}
