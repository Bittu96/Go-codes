package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/segmentio/ksuid"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

const (
	host      = "localhost"
	port      = 5432
	user      = "mrudulkatla"
	password  = "postgres"
	dbname    = "postgres"
	queueName = "fr" //"druid"
)

//type taskLogs struct {
//	ID int `gorm:"type:int;primary_key;auto_increment"`
//	//WorkerID int
//	//TaskID   int
//	Record     string    `gorm:"type:varchar(200)"`
//	RowCreated time.Time `gorm:"type:timestamp;DEFAULT:current_timestamp"`
//}

type taskRecords struct {
	gorm.Model
	WorkerId int           `json:"worker_id" gorm:"int(8)"`
	TaskId   int           `json:"task_id" gorm:"int(8)"`
	Score    int           `json:"score" gorm:"int(16)"`
	Duration time.Duration `json:"duration" gorm:"timestamp"`
	Status   string        `json:"status" gorm:"varchar(16)"`
	Message  string        `json:"message" gorm:"varchar(256)"`
}

type Message2 struct {
	//Data []byte `json:"body"` //without image
	Data       map[string][]byte `json:"body"` //when image is present
	Path       string            `json:"path"`
	Bucket     string            `json:"bucket"`
	XcallID    string            `json:"xcall_id"`
	AgentImage string            `json:"agent_image"`
	StoreID    string            `json:"store_id"`
}

type Mids struct {
	ID            int64     `gorm:"primary_key;auto_increment;not_null"`
	ReqID         string    `gorm:"type:varchar(255);column:req_id"`
	Status        string    `gorm:"type:varchar(255);column:status"`
	Mid           string    `gorm:"type:varchar(255);column:mid"`
	DocCount      int       `gorm:"type:int;column:doc_count"`
	AadhaarCount  int       `gorm:"type:int;column:aadhaar_count"`
	UploadedCount int       `gorm:"type:int;column:uploaded_count"`
	CreatedAt     time.Time `gorm:"DEFAULT:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp" sql:"DEFAULT:NULL" json:"updated_at"`
}

type MidDocs struct {
	ID         int64     `gorm:"primary_key;auto_increment;not_null"`
	XcallID    string    `gorm:"type:varchar(32);primary_key;column:xcall_id"`
	Status     string    `gorm:"type:varchar(255);column:status"`
	Mid        string    `gorm:"type:varchar(255);column:mid"`
	DocID      string    `gorm:"type:varchar(255);column:doc_id"`
	DocType    string    `gorm:"type:varchar(255);column:doc_type"`
	DocFormat  string    `gorm:"type:varchar(255);column:doc_format"`
	Pages      int       `gorm:"type:int;column:pages"`
	Aadhaars   int       `gorm:"type:int;column:aadhaars"`
	Errors     string    `gorm:"type:text;column:errors"`
	CreatedAt  time.Time `gorm:"DEFAULT:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp" sql:"DEFAULT:NULL" json:"updated_at"`
	UploadedAt time.Time `gorm:"type:timestamp" sql:"DEFAULT:NULL" json:"uploaded_at"`
}

type MetaData struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userID"`
	AppID      string    `json:"appID"`
	Path       string    `json:"path"`
	Env        string    `json:"env"`
	MapForAO   string    `json:"MapForAO"`
	Signature  []float64 `json:"signature"`
	Properties struct {
		Mids       interface{} `json:"mids"`
		MidDocs    interface{} `json:"mid_docs"`
		Path       string      `json:"path"`
		StatusCode int         `json:"statusCode"`
		Error      string      `json:"error"`
		ErrorF     string      `json:"errorf"`
		StoreID    string      `json:"store_id"`
		Imagetype  string      `json:"imagetype"` //new for encryption type
		// Face Match
		DC             int     `json:"dc"`
		DCResp         string  `json:"dc_resp"`
		DCRD           string  `json:"dc_rd"`
		DlibF          int     `json:"dlib_f"`
		DlibMP         int     `json:"d_match_per"`
		SufiPercentage float64 `json:"SufiPercentage"`
		SufiScore      float64 `json:"SufiScore"`
		Euclid         float64 `json:"Euclid"`
		DigiFrontNew   int     `json:"digi_front_new"`
		DigiBackNew    int     `json:"digi_back_new"`
		L0             float64 `json:"l0"`
		L1             float64 `json:"l1"`
		L2             float64 `json:"l2"`
		SufiSpan       int     `json:"sufiSpan"`
		UsedRectangle  float64 `json:"usedRect"`
		Distance       float64 `json:"dist"`
		Reason         string  `json:"reason"`
		DlibAutoGreen  float64 `json:"dlib_auto_green"`
		RunFlask       int     `json:"runFlask"`

		DlibPercentage   float64 `json:"dlibPercentage"`
		DlibSpan1        int     `json:"dlibSpan_1"`
		FiopMeanT        float64 `json:"fiopmean_t"`
		FiopMeanAgent    float64 `json:"fiopmeanAgent"`
		FiopMeanAgentAgg float64 `json:"fiopmenAgent_agg"`
		OccThresholdNon  float64 `json:"occ_threshold_non"`
		SufiFlaskSpan    int     `json:"sufiflask_span"`

		// Liveness
		OccSpan          float64 `json:"occSpan"`
		OccMean          float64 `json:"occMean"`
		OccNonReli       float64 `json:"occMean_non_reli"`
		OccThreshold     float64 `json:"occThreshold"`
		OccMask          float64 `json:"occMask"`
		OccMaskThreshold float64 `json:"occMaskThreshold"`

		// OBDC
		PerofAadhar                  float64     `json:"perOfAadhar"`
		RxMin                        float64     `json:"rx_min"`
		RxMax                        float64     `json:"rx_max"`
		RyMin                        float64     `json:"ry_min"`
		RyMax                        float64     `json:"ry_max"`
		MacroScore                   float64     `json:"macro"`
		MacroThreshold               float64     `json:"macro_t"`
		MacroBackNew                 float64     `json:"macro_new"`
		MacroBackNewThreshold        float64     `json:"macro_new_t"`
		AadharFrontSpoof             float64     `json:"afs"`
		AadharFrontSpoofThreshold    float64     `json:"afs_t"`
		AadharFrontSpoofNew          float64     `json:"afs_new"`
		AadharFrontSpoofThresholdNew float64     `json:"afs_new_t"`
		AadharBackSpoof              float64     `json:"abs"`
		AadharBackSpoofThreshold     float64     `json:"abs_t"`
		AadharBackSpoofNew           float64     `json:"abs_new"`
		AadharBackSpoofThresholdNew  float64     `json:"abs_new_t"`
		AadharFrontDigiSpoof         float64     `json:"afds_new"`
		AadharFrontPhotoSpoof        float64     `json:"afps_new"`
		AadharFrontRedLine           float64     `json:"microRedLine"`
		AadharFrontRedLineThreshold  float64     `json:"microRedLine_t"`
		AadharFrontBorderSpoof       float64     `json:"af_border_s"`
		Af_border_obd_s              float64     `json:"af_border_obd_s"`
		AadharFrontBorderSpoofThres  float64     `json:"af_border_t"`
		AadharFrontPaperSpoof        float64     `json:"afmps_new"`
		AadharFrontPaperSpoofThres   float64     `json:"afmps_new_t"`
		Doublecard                   float64     `json:"double_card"`
		DoublecardThreshold          float64     `json:"double_card_t"`
		PaperSpoofSpan               int         `json:"paperSpoofSpan"`
		ASLgr                        float64     `json:"ASLgr"`
		AbsLgr                       float64     `json:"AbsLgr"`
		AadharNumSpoofPOA            float64     `json:"abns_new"`
		AadharNumSpoofPOI            float64     `json:"afns_new"`
		ObdcData                     string      `json:"obdc_data"`
		MicroScore                   float64     `json:"micro"`
		MicroThreshold               float64     `json:"micro_t"`
		BlurScore                    float64     `json:"blr"`
		BlurThreshold                float64     `json:"blr_t"`
		IsBack                       string      `json:"is_back"`
		IsBackClient                 string      `json:"is_back_client"`
		AbnormalityIndex             string      `json:"ab_index"`
		TiltAngle                    float64     `json:"t_angle"`
		TiltThreshold                float64     `json:"t_threshold"`
		BlackAndWhite                float64     `json:"bw"`
		BlackAndWhiteThreshold       float64     `json:"bw_t"`
		SpoofFiopSpan                string      `json:"spooffiopspan"`
		CheckBlurFacesSpan           int         `json:"checkBlurFacesSpan"`
		CheckMultipleFacesSpan       int         `json:"checkMultipleFacesSpan"`
		TotalSpan                    int         `json:"latency"`
		EncodeSpan                   int         `json:"encodeSpan"`
		EyeSpan                      int         `json:"eyeSpan"`
		QrSpan                       int         `json:"qrSpan"`
		AadharBackSpan               int         `json:"aadharBackSpan"`
		DkycSpan                     int         `json:"dkycSpan"`
		EyesSpan                     int         `json:"eyesSpan"`
		BlurSpan                     int         `json:"blurnessSpan"`
		BlackWhiteSpan               int         `json:"BlackWhiteSpan"`
		AadharFrontSpan              int         `json:"aadharFrontSpan"`
		WBSPan                       int         `json:"wbSpan"`
		StartTime                    string      `json:"starttime"`
		EndTime                      string      `json:"endtime"`
		FrThreshold                  float64     `json:"FrThreshold"`
		FiopMean                     float64     `json:"fiopMean"`
		FiopMeanAggresive            float64     `json:"fiopMean_agg"`
		FiopMeanAggThreshold         float64     `json:"fiopMean_agg_t"`
		SigmoidAgg                   float64     `json:"sigmoid_agg"`
		Confidence                   int         `json:"confidence"`
		Result                       interface{} `json:"result"`
		IsAadhar                     string      `json:"is_aadhar"`
		IsQrCode                     string      `json:"is_qr_code"`
		QrResult                     string      `json:"qr_result"`
		OcrResult                    string      `json:"ocr_result"`
		IsAadharReal                 string      `json:"is_aadhar_real"`
		OcrBlock                     string      `json:"ocr_block"`
		OcrPlots                     string      `json:"ocr_plots"`
		Ocrplots                     interface{} `json:"ocrplots"`
		WhiteBackground              string      `json:"white_background"`
		EyesOpen                     string      `json:"eyes_open"`
		IsRectangle                  string      `json:"is_rectangle"`
		IsRectCust                   string      `json:"is_rect_cust"`
		AadharFrontPartial           int         `json:"aadharFrontPartialSpan"`
		CheckAadharBackSpoof         int         `json:"checkAadharBackSpoof"`
		DeviceType                   string      `json:"device_type"`
		TemplateData                 string      `json:"template_data"`
		Plot_check_combinedpoa       string      `json:"plot_check_combinedpoa"`
		CompleteDatesText            string      `json:"completeDatesText"`
		OcrTextCombined              string      `json:"ocr_text_combined"`
		OCRCombinedPOAOnly           string      `json:"ocr_combined_poa_only"`
		OcrTextCombinedPOASpan       int         `json:"ocrTextCombinedPOASpan"`
		Ocr_text_poi                 interface{} `json:"ocr_text_poi"`
		ClientAadharModel            string      `json:"clientAadharModel"`
		ClientFaceModel              string      `json:"clientFaceModel"`
		DocTypeClient                string      `json:"doc_type_client"`
		ClientaadharmodelError       string      `json:"clientAadharModel_Error"`
		ClientfacemodelError         string      `json:"clientFaceModel_Error"`
		Addressconf                  int         `json:"addressconf"`
		Filenoconf                   int         `json:"filenoconf"`
		Opassconf                    int         `json:"opassconf"`
		Opoiconf                     int         `json:"opoiconf"`
		Odoiconf                     int         `json:"odoiconf"`
		Fatherconf                   int         `json:"fatherconf"`
		Motherconf                   int         `json:"motherconf"`
		Spouseconf                   int         `json:"spouseconf"`
		Ageconf                      int         `json:"ageconf"`
		Dateconf                     int         `json:"dateconf"`
		Dobconf                      int         `json:"dobconf"`
		Genderconf                   int         `json:"genderconf"`
		Voteridconf                  int         `json:"voteridconf"`
		Nameconf                     int         `json:"nameconf"`
		Relationconf                 int         `json:"relationconf"`
		Voter_card_type              string      `json:"voter_card_type"`

		//testing purposes , delete in production
		TotalPlots            int         `json:"totalPlots"`
		TotalPlotsCombinedPOA int         `json:"totalPlotsCombinedPOA"`
		IsFront               string      `json:"is_front"`
		IsLive                string      `json:"is_live"`
		IsAadharDig           string      `json:"is_aadhar_dig"`
		IsFrontOCR            string      `json:"is_front_ocr"`
		IsBackOcr             string      `json:"is_back_ocr"`
		IsFaceDetected        string      `json:"is_face_detected"`
		IsWhiteBack           string      `json:"is_white_back"`
		GvDetails             interface{} `json:"gvDetails"`
		GossDetails           interface{} `json:"gossDetails"`
		GvOCRBlock            string      `json:"gvOcrBlock"`
		GossOCRBlock          string      `json:"gossOcrBlock"`
		CoOrdinates           string      `json:"coordinates"`
		Boxes                 int         `json:"Boxes"`
		MultipleFaces         float64     `json:"multipleFaces"`
		OcrSpan               int         `json:"ocrTextSpan"`
		OcrEATextSpan         int         `json:"ocrEATextSpan"`
		DatesflaskSpan        int         `json:"datesflaskSpan"`
		SocrSpan              int         `json:"socrSpan"`
		RotationSpan          int         `json:"RotationSpan"`
		ResizeSpan            int         `json:"ResizeSpan"`
		Getocrspan            int         `json:"Getocrspan"`
		AadharDoubleCardSpan  int         `json:"aadharDoubleCardSpan"`
		BlurFaces             float64     `json:"checkBlurFaces"`
		BlurFacesThreshold    float64     `json:"checkBlurFaces_t"`
		LiveType              string      `json:"capture_type"`
		MatchType             string      `json:"match_type"`
		Delay                 string      `json:"delay"`
		OriginalAddress       string      `json:"Original_Address"`
		FaceMatchThreshold    float64     `json:"fm_t"`
		//white bg
		WhiteValue     float64 `json:"wb_s"`
		WhiteThreshold float64 `json:"wb_t"`
		WhiteBgScore   float64 `json:"wbm_s"`
		Mask           string  `json:"Mask"`
		WhiteVal       string  `json:"White_Value"`
		WHiteValCrops  int     `json:"White_Value_Crops"`
		Crop           string  `json:"Crops"`
		Vectors        string  `json:"vectors"`
		Version        string  `json:"version"`
		StoreId        string  `json:"storeId"`

		//global blacklisting Aadhar
		AadharMaxDist          float64 `json:"aadhar_max_dist"`
		AadharMaxUserId        string  `json:"aadhar_max_userid"`
		AadharUpperThreshold   float64 `json:"aadhar_bl_upper_thres"`
		AadharLowerThreshold   float64 `json:"aadhar_bl_lower_thres"`
		UpperThreshold         string  `json:"is_upper_thres"`
		LowerThreshold         string  `json:"is_lower_thres"`
		DocFaceThreshold       string  `json:"is_docface_thres"`
		GlobalBLaadhar         string  `json:"is_globalBL_aadhar"`
		AadharDocFaceThreshold float64 `json:"aadhar_docface_thres"`
		AadharDocfaceDist      float64 `json:"gl_doc_face_dist"`
		AadharMatchId          string  `json:"aadhar_match_id"`

		//Agent Global Blacklisitng
		Agentmax_dist      float64 `json:"agent_bl_max_dist"`
		AgentMaxUserId     string  `json:"agent_bl_max_userid"`
		AgentMaxMatchId    string  `json:"agent_bl_match_id"`
		AgentBlThreshold   float64 `json:"agent_global_bl_thres"`
		IsAgentGlobalBl    string  `json:"is_agent_globalBL"`
		ISagentWhitelisted string  `json:"is_agent_whitelisted"`

		// OCR Front
		OCRFullText          string  `json:"ocr_full_text"`
		Plot_check           string  `json:"plot_check"`
		TrustOcr             float64 `json:"trust_ocr"`
		SocrName             string  `json:"Name"`
		SocrUID              string  `json:"UID"`
		SocrDOB              string  `json:"DOB"`
		SocrYob              string  `json:"YOB"`
		SocrGender           string  `json:"Gender"`
		SocrFather           string  `json:"Father"`
		GvName               string  `json:"NameGV"`
		GvUID                string  `json:"UIDGV"`
		GvDOB                string  `json:"DOBGV"`
		GvYob                string  `json:"YOBGV"`
		GvGender             string  `json:"GenderGV"`
		GvFather             string  `json:"FatherGV"`
		OcrTextClient        string  `json:"ocr_text_client"`
		PoiOcrConf           int     `json:"poiOcrConf"`
		PoaOcrConf           int     `json:"poaOcrConf"`
		IsSpecularityFound   string  `json:"isSpecularityFound"`
		Name_conf            int     `json:"name_conf"`
		Father_conf          int     `json:"father_conf"`
		Mother_conf          int     `json:"mother_conf"`
		Dob_conf             int     `json:"dob_conf"`
		Yob_conf             int     `json:"yob_conf"`
		Gender_conf          int     `json:"gender_conf"`
		SpecularityFlaskSpan int     `json:"SpecularityFlaskSpan"`
		Uid_conf             int     `json:"uid_conf"`

		//OCR Back
		SocrPincode    string `json:"pincode"`
		SocrCareOf     string `json:"care_of"`
		SocrAddharNum  string `json:"AadharNumber"`
		SocrAddValue   string `json:"Address_value"`
		SocrLocality   string `json:"locality"`
		SocrDistrict   string `json:"District"`
		SocrState      string `json:"State"`
		SocrCity       string `json:"city"`
		SocrStreetName string `json:"streetName"`
		SocrLandmark   string `json:"landmark"`
		SocrHouseNum   string `json:"houseNum"`

		GvPincode    string `json:"pincodeGV"`
		GvCareOf     string `json:"care_ofGV"`
		GvAddharNum  string `json:"AadharNumberGV"`
		GvAddValue   string `json:"Address_valueGV"`
		GvLocality   string `json:"localityGV"`
		GvDistrict   string `json:"DistrictGV"`
		GvState      string `json:"StateGV"`
		GvCity       string `json:"cityGV"`
		GvStreetName string `json:"streetNameGV"`
		GvHouseNum   string `json:"houseNumGV"`
		GvLandmark   string `json:"landmarkGV"`

		//lgr
		PersonProb float64 `json:"personProb"`
		ScreenProb float64 `json:"screenProb"`
		PrintProb  float64 `json:"printProb"`
		Sigmoid    float64 `json:"sigmoid"`

		//adhar
		AreaOfleftSide   float64 `json:"areaOfleftSide"`
		AreaOfrightSide  float64 `json:"areaOfrightSide"`
		AreaOfBottomSide float64 `json:"areaOfBottomSide"`
		AreaOftopSIde    float64 `json:"areaOftopSIde"`

		//Redis Keys
		RedisGetSpan  int    `json:"redisGetSpan"`
		RejectByRedis string `json:"rej_by_redis"`
		RedisSetSpan  int    `json:"redisSetSpan"`

		// AO Dashboard
		MapForAO interface{} `json:"mapforao"`

		//document type and journey type
		Journey     string `json:"journey"`
		Document    string `json:"document"`
		DocType     string `json:"doc_type"`
		CircleId    string `json:"circle_id"`
		Masterid    string `json:"master_id"`
		AgentId     string `json:"agent_id"`
		OrderInfo   string `json:"order_info"`
		CafNumber   string `json:"caf_number"`
		GtRr        string `json:"gt_rr"`
		DeviceModel string `json:"device_model"`
		TrailCount  string `json:"trail_count"`

		//cpu
		CpuRedline            float64 `json:"cpu_redline"`
		CpuMacroBack          float64 `json:"cpu_macro_back"`
		CpuMacroBackNew       float64 `json:"cpu_macro_back_new"`
		CpuSpoofFront         float64 `json:"cpu_afs"`
		CpuBlur               float64 `json:"cpu_blr"`
		CpuMacroFront         float64 `json:"cpu_macro_front"`
		CpuMicroFront         float64 `json:"cpu_micro_front"`
		CpuBackSpoof          float64 `json:"cpu_abs_new"`
		CpuFiop               float64 `json:"cpu_fiop"`
		CpuFiopAgg            float64 `json:"cpu_fiop_agg"`
		CpuFiopV2             float64 `json:"cpu_fiop_v2"`
		CpuBlurFaces          float64 `json:"cpu_blur_faces"`
		CpuMultipleFaces      float64 `json:"cpu_multiple"`
		CpuWhiteBg            float64 `json:"cpu_wb"`
		CpuOcclusions         float64 `json:"cpu_occ"`
		CpuOcclusionMask      float64 `json:"cpu_occmask"`
		CpuLivenesspaperspoof float64 `json:"cpu_psl"`
		CpuBorderSpoof        float64 `json:"cpu_af_border_s"`
		CpuDoubleCard         float64 `json:"cpu_double_card"`
		QrFlag                int     `json:"flag"`

		// NON Aadhar
		PanValue                    float64 `json:"pan"`
		PanThreshold                float64 `json:"pan_t"`
		PanMicro                    float64 `json:"pan_micro"`
		PanMicroThreshold           float64 `json:"pan_micro_t"`
		VoterFrontMacro             float64 `json:"voter_f_mac"`
		VoterFrontMacroThreshold    float64 `json:"voter_f_mac_t"`
		VoterBackMacro              float64 `json:"voter_b_mac"`
		VoterBackMacroThreshold     float64 `json:"voter_f_mac_t"`
		PassportBackMacro           float64 `json:"passport_b_mac"`
		PassportBackMacroThreshold  float64 `json:"passport_b_mac_t"`
		PassportFrontMacro          float64 `json:"passport_f_mac"`
		PassportFrontMacroThreshold float64 `json:"passport_f_mac_t"`
		PassportFrontMicro          float64 `json:"passport_f_mic"`
		PassportFrontMicroThreshold float64 `json:"passport_f_mic_t"`
		VoterFrontMicro             float64 `json:"voter_f_mic"`
		VoterFrontMicroThreshold    float64 `json:"voter_f_mic_t"`
		Eam_s                       float64 `json:"eam_s"`
		Eam_t                       float64 `json:"eam_t"`
		BlackStore                  string  `json:"blackstore"`
		Apptype                     string  `json:"app_type"`
		RandomId                    string  `json:"random_id"`
		Lat                         string  `json:"lat"`
		Long                        string  `json:"long"`
		//charConf
		WordCharMapPOI interface{} `json:"OCRMapPoI"`
		WordCharMapPOA interface{} `json:"OCRMapPoA"`

		//Anamoly Models
		WhiteBGAnamolyScore float64 `json:"white_bg_anamoly_score"`

		//Dedupe Distance
		Max_UserId string  `json:"max_userid"`
		Max_Dist   float64 `json:"max_dist"`

		//Face Detection Model
		Flag1                       string      `json:"flag1"`
		Flag2                       string      `json:"flag2"`
		DlibFlag1                   string      `json:"dlib_flag1"`
		DlibFlag2                   string      `json:"dlib_flag2"`
		EyeScoreRight               float64     `json:"eyeScore_right"`
		EyeScoreLeft                float64     `json:"eyeScore_left"`
		DlibImg1Enc                 []float64   `json:"dlib_img1_encodings"`
		SufiImg1Enc                 []float64   `json:"sufi_img1_encodings"`
		SufiImg2Enc                 []float64   `json:"sufi_img2_encodings"`
		SufiEnc                     []float64   `json:"sufi_enc"`
		AgentSufienc                []float64   `json:"agent_sufi_enc"`
		SufiDist                    float64     `json:"sufi_dist"`
		DlibEncodings               []float64   `json:"dlib_encodings"`
		CheckBlackListSignatures    string      `json:"check_BlackListSignatures"`
		BlacklistSignatureDist      float64     `json:"blacklistSignature_dist"`
		CheckBlacklistAgent         string      `json:"check_Blacklist_Agent"`
		Isblacklist                 string      `json:"is_blacklist"`
		IsblacklistRIL              string      `json:"is_blacklist_RIL"`
		CheckBlacklistCustomer      string      `json:"check_Blacklist_Cust"`
		CheckAgentBlSignature       string      `json:"check_Agent_bl_signature"`
		BLDlibDist                  float64     `json:"bl_dlib_dist"`
		BLThresholdChange           string      `json:"bl_thresh_change"`
		BLFiopThresholdChange       string      `json:"bl_threshold"`
		IsCustomerFiopModel         string      `json:"is_cust_models"`
		Bl_Fiop_Threshold           float64     `json:"bl_fiop_threshold"`
		Bl_Fiop2_Threshold          float64     `json:"bl_fiop2_threshold"`
		Bl_FiopAgg_Threshold        float64     `json:"bl_fiopAgg_threshold"`
		Bl_Fiop_PS_Threshold        float64     `json:"bl_fiop_ps_threshold"`
		FiopCustV2                  float64     `json:"fiopMean_v2"`
		FiopCustV2Thres             float64     `json:"fiopMean_v2_t"`
		FiopAgentScore              float64     `json:"fiopMean_Agent"`
		FiopMeanAgent_t             float64     `json:"fiopMean_Agent_t"`
		FiopAgentAggScore           float64     `json:"fiopMean_Agent_agg"`
		FiopMeanAgent_Agg_t         float64     `json:"fiopMean_Agent_agg_t"`
		Paperspoofliveness_t        float64     `json:"fiop_ps_t"`
		Paperspoofliveness          float64     `json:"fiop_ps"`
		IsPaperSpoof                string      `json:"is_paperspoof"`
		CheckPSLivenessSpan         int         `json:"checkPSLiveness"`
		Is_agent_agg                string      `json:"is_agent_agg"`
		ObdcLeft                    float64     `json:"obdc_left"`
		ObdcRight                   float64     `json:"obdc_right"`
		ObdcTop                     float64     `json:"obdc_top"`
		ObdcBottom                  float64     `json:"obdc_bottom"`
		AadhaarMicroOBDFailed       string      `json:"AadhaarMicroOBDFailed"`
		MicroPixthreshold           float64     `json:"microPixthreshold"`
		UseAadharMicroOBD           string      `json:"UseAadharMicroOBD"`
		PhotoSpoofThreshold         float64     `json:"psf_t"`
		PhotoSpoof                  float64     `json:"psf_s"`
		DigiSpoofThreshold          float64     `json:"dsf_t"`
		DigiSpoof                   float64     `json:"dsf_s"`
		PanClothPaperThreshold      float64     `json:"pan_clother_paper_t"`
		PanClothPaper               float64     `json:"pan_clother_paper"`
		PanClothMidPaper            float64     `json:"pan_clother_mid_paper"`
		PanClothGenPaper            float64     `json:"pan_clother_gen_paper"`
		PassportClothPaperThreshold float64     `json:"passport_clother_paper_t"`
		PassportClothPaper          float64     `json:"passport_clother_paper"`
		PassportClothMidPaper       float64     `json:"passport_clother_mid_paper"`
		PassportClothGenPaper       float64     `json:"passport_clother_gen_paper"`
		VoterClothPaperThreshold    float64     `json:"voter_clother_paper_t"`
		VoterClothPaper             float64     `json:"voter_clother_paper"`
		VoterClothPaperBW           float64     `json:"voter_clother_paper_bw"`
		VoterClothMidPaperBW        float64     `json:"voter_clother_mid_paper_bw"`
		VoterClothGenPaperBW        float64     `json:"voter_clother_gen_paper_bw"`
		VoterClothPaperColor        float64     `json:"voter_clother_paper_color"`
		VoterClothMidPaperColor     float64     `json:"voter_clother_mid_paper_color"`
		VoterClothGenPaperColor     float64     `json:"voter_clother_gen_paper_color"`
		AadhaarClothPaperThreshold  float64     `json:"aadhaar_clother_paper_t"`
		AadhaarClothPaper           float64     `json:"aadhaar_clother_paper"`
		AadhaarClothMidPaper        float64     `json:"aadhaar_clother_mid_paper"`
		AadhaarClothGenPaper        float64     `json:"aadhaar_clother_gen_paper"`
		A4ClothPaperThreshold       float64     `json:"a4_clother_paper_t"`
		A4ClothPaper                float64     `json:"a4_clother_paper"`
		CombineClothPaperThreshold  float64     `json:"combine_clother_paper_t"`
		CombineClothPaper           float64     `json:"combine_clother_paper"`
		SmartAadhar                 float64     `json:"smart_aadhar"`
		SmartAadharThreshold        float64     `json:"smart_aadhar_t"`
		SmartAadharFound            string      `json:"SmartAadharFound"`
		BlRefId                     string      `json:"bl_ref_id"`
		VoterPoi_1                  float64     `json:"voter_poi_1"`
		VoterPoi_2                  float64     `json:"voter_poi_2"`
		VoterPoi_3                  float64     `json:"voter_poi_3"`
		VoterPoa                    float64     `json:"voter_poa"`
		DistMistake                 string      `json:"dist_mistake"`
		SubstrFound                 string      `json:"substr_found"`
		VoterPSfrontBWThreshold     float64     `json:"voter_ps_f_bw_t"`
		VoterPSfrontBW              float64     `json:"voter_ps_f_bw"`
		VoterPSfrontColorThreshold  float64     `json:"voter_ps_f_color_t"`
		VoterPSfrontColo            float64     `json:"voter_ps_f_color"`
		Cords                       interface{} `json:"cords"`
		PassportbspoofThreshold     float64     `json:"passport_b_spoof_t"`
		Passportbspoof              float64     `json:"passport_b_spoof"`
		PassportfspoofThreshold     float64     `json:"passport_f_spoof_t"`
		Passportfspoof              float64     `json:"passport_f_spoof"`
		Ocrrect                     interface{} `json:"ocrrect"`
		Poikey                      string      `json:"poi_key"`
		Redis_Agent_Non_BL_getvalue int         `json:"Redis_Agent_Non_BL_getvalue"`
		Redis_Agent_Non_BL_setvalue int         `json:"Redis_Agent_Non_BL_setvalue"`
		Redis_Agent_BL_getvalue     int         `json:"Redis_Agent_BL_getvalue"`
		Redis_Agent_BL_setvalue     int         `json:"Redis_Agent_BL_setvalue"`
		Redis_Cust_Non_BL_getvalue  int         `json:"Redis_Cust_Non_BL_getvalue"`
		Redis_Cust_Non_BL_setvalue  int         `json:"Redis_Cust_Non_BL_setvalue"`
		Redis_Cust_BL_getvalue      int         `json:"Redis_Cust_BL_getvalue"`
		Redis_Cust_BL_setvalue      int         `json:"Redis_Cust_BL_setvalue"`

		WhiteBGAnomalyScore  float64 `json:"white_bg_anomaly_score"`
		LivenessAnomalyScore float64 `json:"liveness_anomaly_score"`
	} `json:"properties"`
}

var dbConn *gorm.DB

func GetConnection() *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)

	if err != nil {
		log.Println(err)
		panic("failed to connect database")
	}

	log.Println("DB Connection established...")
	return db
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func send() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "Hello World!"
	fmt.Println(ctx, q)

	// test message for rpsl
	xcallId := ksuid.New().String()

	var testMetaData MetaData
	testMids := Mids{Mid: "testmidid", ReqID: xcallId}
	testMidDocs := MidDocs{Mid: "testmiddocid", XcallID: xcallId}

	testMetaData.ID = xcallId
	testMetaData.Signature = []float64{1.1, 2.2, 3.3}
	testMetaData.Properties.Mids = testMids
	testMetaData.Properties.MidDocs = testMidDocs

	metaDataBytes, _ := json.Marshal(testMetaData)
	log.Println("xcallid:", xcallId)
	testMessage := Message2{XcallID: xcallId, Data: map[string][]byte{"meta.json": metaDataBytes}}
	msgBytes, _ := json.Marshal(testMessage)

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msgBytes, //[]byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}

func receive(dbConn *gorm.DB) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		var iter int
		for d := range msgs {
			iter++
			log.Printf("Received message %v: %s", iter, d.Body)
			wg.Add(1)
			go InsertInDB(wg, dbConn, d.Body)
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	wg.Wait()
	log.Printf(" [*] Wait is over!")
	<-forever
}

func InsertInDB(wg *sync.WaitGroup, dbConn *gorm.DB, record []byte) {
	// var taskRecord = taskLogs{Record: string(record), RowCreated: time.Now()}
	var taskRecord taskRecords
	uErr := json.Unmarshal(record, &taskRecord)
	if uErr != nil {
		log.Println("uErr : ", uErr)
		panic(uErr)
	}
	//
	err := dbConn.Create(&taskRecord).Error
	if err != nil {
		log.Println("log not created", err)
	} else {
		log.Println("log created", err)
	}
	wg.Done()
}

func main() {
	send()

	//dbConn := GetConnection()
	//wg := &sync.WaitGroup{}
	//InsertInDB(wg, dbConn, message)
	//receive(dbConn)
}
