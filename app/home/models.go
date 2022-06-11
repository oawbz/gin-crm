package home

type menuMode struct {
	//菜单
	Name   string `form:"name" json:"name" binding:"required"`
	Remark string `form:"remark" json:"remark" binding:"required"`
	Data   string `form:"data" json:"data" binding:"required"`
}

type navigationMode struct {
	//导航
	Label     string `column:"label" json:"label" type:"string"`
	Icon      string `column:"icon" json:"icon" type:"string"`
	Url       string `column:"url" json:"url"  type:"string"`
	Redirect  string `column:"redirect" json:"redirect"  type:"string"`
	SchemaApi string `column:"schemaApi" json:"schemaApi"  type:"string"`
}
type projectAdd struct { //项目添加
	Explain      string `column:"explain" json:"explain" type:"string"`
	Name         string `column:"name" json:"name" type:"string"`
	Addtime      int64  `column:"addtime" json:"addtime" type:"string"`
	Director     int64  `column:"director" json:"director" type:"int"`
	Platformlist string `platformlist:"name" json:"platformlist" type:"string"`
	Doctorlist   string `column:"doctorlist" json:"doctorlist" type:"string"`
}

type platformAdd struct { //平台添加
	Explain string `column:"explain" json:"explain" type:"string"`
	Name    string `column:"name" json:"name" type:"string"`
	Addtime int64  `column:"addtime" json:"addtime" type:"string"`
	Uid     int64  `column:"uid" json:"uid" type:"int"`
}
type userAdd struct { //用户添加

	Name    int64  `column:"name" json:"name" type:"int"`
	Pas     string `column:"pas" json:"pas" type:"string"`
	Nick    string `column:"nick" json:"nick" type:"string"`
	Allow   int64  `column:"allow" json:"allow" type:"int"`
	Group   int64  `column:"group" json:"group" type:"int"`
	Addtime int64  `column:"addtime" json:"addtime" type:"int"`
}
type doctorAdd struct { //医生添加

	Name      string `column:"name" json:"name" type:"string"`
	Group     int64  `column:"group" json:"group" type:"int"`
	Monday    int64  `column:"monday" json:"monday" type:"int"`
	Tuesday   int64  `column:"tuesday" json:"tuesday" type:"int"`
	Wednesday int64  `column:"wednesday" json:"wednesday" type:"int"`
	Thursday  int64  `column:"thursday" json:"thursday" type:"int"`
	Friday    int64  `column:"friday" json:"friday" type:"int"`
	Saturday  int64  `column:"saturday" json:"saturday" type:"int"`
	Sunday    int64  `column:"sunday" json:"sunday" type:"int"`
	Addtime   int64  `column:"addtime" json:"addtime" type:"int"`
}

type list struct { //通用列表
	Count int64       `column:"count" json:"count" `
	Rows  interface{} `column:"rows" json:"rows" `
}

type manageroptions struct { //项目主管选项列表
	Nick string `column:"nick" json:"label" type:"string"`
	Id   int64  `column:"id" json:"value" type:"int"`
}
type platformoptions struct { //平台\医生\权限选项列表
	Name string `column:"name" json:"label" type:"string"`
	Id   int64  `column:"id" json:"value" type:"int"`
}

type delstate struct { //删除结果
	Status int64  `column:"status" json:"status" `
	Msg    string `column:"msg" json:"msg" `
}

type password struct {
	//修改密码
	Pas1 string `form:"password" json:"password"`
	Pas2 string `form:"password1" json:"password1"`
	Pas3 string `form:"password2" json:"password2"`
}

type dataEntryAdd struct { //数据录入
	Phone     int64  `column:"phone" json:"phone" type:"int"`
	Name      string `column:"name" json:"name" type:"string"`
	Gender    int64  `column:"gender" json:"gender" type:"int"`
	Age       int64  `column:"age" json:"age" type:"int"`
	Doctor    int64  `column:"doctor" json:"doctor" type:"int"`
	Illness   string `column:"illness" json:"illness" type:"string"`
	Platform  int64  `column:"platform" json:"platform" type:"int"`
	Phoneuser int64  `column:"phoneuser" json:"phoneuser" type:"int"`
	Project   int64  `column:"project" json:"project" type:"int"`
	Addtime   int64  `column:"addtime" json:"addtime" type:"int"`
}
type dataEntryRenew struct { //数据修改
	Name     string `column:"name" json:"name" type:"string"`
	Card     string `column:"card" json:"card" type:"string"`
	Gender   int64  `column:"gender" json:"gender" type:"int"`
	Age      int64  `column:"age" json:"age" type:"int"`
	Doctor   int64  `column:"doctor" json:"doctor" type:"int"`
	Illness  string `column:"illness" json:"illness" type:"string"`
	Platform int64  `column:"platform" json:"platform" type:"int"`
}
type dictionaryUserList struct { //当前组用户列表
	Nick string `column:"nick" json:"label" type:"string"`
	Id   int64  `column:"id" json:"value" type:"int"`
}
type entryPhone struct { //添加电话
	Id       int64  `column:"id" json:"id" type:"int"`
	Phone    string `column:"phone" json:"phone" type:"string"`
	Name     string `column:"name" json:"name" type:"string"`
	Platform int64  `column:"platform" json:"platform" type:"int"`
	Addtime  int64  `column:"addtime" json:"addtime" type:"int"`
}
type distributionup struct { //数据分配
	Consultfp int64  `column:"consultfp" json:"consultfp" type:"int"`
	Ids       string `column:"ids" json:"ids" type:"string"`
}
type returnVisitadd struct { //回访内容

	Content     string `column:"content" json:"content" type:"string"`
	Customerid  string `column:"customerid" json:"customerid" type:"int"`
	Revisitdays int64  `column:"revisitdays" json:"revisitdays" type:"int"`
	Result      int64  `column:"result" json:"result" type:"int"`
	User        int64  `column:"user" json:"user" type:"string"`
}
type zxReserve struct { //咨询预约

	Customerid      string `column:"customerid" json:"customerid" type:"string"`
	Projectid       int64  `column:"projectid" json:"projectid" type:"int"`
	TreatmentMethod int64  `column:"treatment_method" json:"treatment_method" type:"int"`
	Doctor          int64  `column:"doctor" json:"doctor" type:"int"`
	VisitTime       string `column:"visit_time" json:"visit_time" type:"string"`
	RegistrationFee int64  `column:"registration_fee" json:"registration_fee" type:"int"`
	PayTime         string `column:"pay_time" json:"pay_time" type:"string"`
	PatientRating   string `column:"patient_rating" json:"patient_rating" type:"string"`
	Remarks         string `column:"remarks" json:"remarks" type:"string"`
	Addtime         int64  `column:"addtime" json:"addtime" type:"int"`
	User            int64  `column:"user" json:"user" type:"int"`
	Repeats         int64  `column:"repeats" json:"repeats" type:"int"`
}
type dzReserve struct { //到诊

	Hsid            string `column:"hsid" json:"hsid" type:"string"`
	TreatmentMethod int64  `column:"treatment_method" json:"treatment_method" type:"int"`
	Drug_cost       int64  `column:"drug_cost" json:"drug_cost" type:"int"`
	Frequency       int64  `column:"frequency" json:"frequency" type:"int"`
	ReagentType     int64  `column:"reagent_type" json:"reagent_type" type:"int"`
	ProcessCost     int64  `column:"process_cost" json:"process_cost" type:"int"`
	Doctor          int64  `column:"doctor" json:"doctor" type:"int"`
	VisitTime       string `column:"visit_time" json:"visit_time" type:"string"`
	RegistrationFee int64  `column:"registration_fee" json:"registration_fee" type:"int"`
	PayTime         string `column:"pay_time" json:"pay_time" type:"string"`
	PatientRating   string `column:"patient_rating" json:"patient_rating" type:"string"`
	Remarks         string `column:"remarks" json:"remarks" type:"string"`
	Repeats         int64  `column:"repeats" json:"repeats" type:"int"`
	State           int64  `column:"state" json:"state" type:"int"`
}

type statisticsdata struct { //统计日报
	Name    string `column:"name" json:"name" type:"string"`
	Phone   int    `column:"phone" json:"phone" type:"int"`
	Reserve int    `column:"reserve" json:"reserve" type:"int"`
	Rate    string `column:"rate" json:"rate" type:"string"`
}

type yuedata struct { //统计月报
	Name    string `column:"name" json:"name" type:"string"`
	Phone   int    `column:"phone" json:"phone" type:"int"`
	Reserve int64  `column:"reserve" json:"reserve" type:"int"`
	Arrive  int64  `column:"arrive" json:"arrive" type:"string"`
	Amount  int64  `column:"amount" json:"amount" type:"string"`
}

type Body struct {
	Type     string `json:"type" type:"string"`
	Name     string `json:"name" type:"string"`
	Label    string `json:"label" type:"string"`
	Lequired bool   `json:"required" type:"bool"`
	Disabled bool   `json:"disabled" type:"bool"`
	Value    int64  `json:"value" type:"int"`
	//Formula  string `json:"formula" type:"string"`
}
