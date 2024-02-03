package chaincodeTranscript

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ------------------------------------------------------------------------------------------------------
// *
// *				SAMPLES (CLI) - INTERACTIONS WITH THE DEPLOYED SMART CONTRACT
// *
// ------------------------------------------------------------------------------------------------------

// 1- To initialize the ledger with a student records: one StudentInfo, eight TakenCourse, and eight CourseInfo records
// peer chaincode invoke -C mychannel -n mySmartContract -c '{"function":"InitLedger","Args":[]}'

// 2- To query whether a record already exists or not
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["IsRecordExists", "Fenerbahce University", "190908809", "48c4c683034af0c0a03fbda1d9a1f7cd"]}'

// 3- To create new records from student information (StudentInfo), course information (CourseInfo), and results of courses achieved by a student (TakenCourse)
// peer chaincode invoke -C mychannel -n mySmartContract -c '{"function":"InsertNewRecordStudentInfo","Args":["Fenerbahce University", "Faculty of Engineering and Architecture", "Department of Computer Engineering", "299799009", "Selvi", "Ahmet", "44262495576", "02.09.2022", "Major / OSYM", "Undergraduate", "2", "3"]}'
// peer chaincode invoke -C mychannel -n mySmartContract -c '{"function":"InsertNewRecordTakenCourse","Args":["Fenerbahce University", "299799009", "COMP2004", "BB", "18", "4"]}'
// peer chaincode invoke -C mychannel -n mySmartContract -c '{"function":"InsertNewRecordCourseInfo","Args":["Fenerbahce University", "299799009", "COMP2004", "Database Management Systems", "C", "6", "3"]}'

// 4- To query for fetching a student's relevant records from Hyperledger Fabric CouchDB
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["Get_Student_StudentInfo", "Fenerbahce University", "190908809"]}'
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["Get_Student_CourseInfos", "Fenerbahce University", "190908809"]}'
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["Get_Student_TakenCourses", "Fenerbahce University", "190908809"]}'

// 5- To query for fetching a higher education institution's relevant records from Hyperledger Fabric CouchDB
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["Get_HEI_TakenCourses", "Fenerbahce University"]}'
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["Get_HEI_CourseInfos", "Fenerbahce University"]}'
// peer chaincode query -C mychannel -n mySmartContract -c '{"Args":["Get_HEI_StudentInfos", "Fenerbahce University"]}'

// 6- To query for fetching a student's transcript from Hyperledger Fabric CouchDB
// peer chaincode invoke -C mychannel -n mySmartContract -c '{"function":"GetStudentTranscript","Args":["Fenerbahce University", "190908809"]}'

// ------------------------------------------------------------------------------------------------------
// *
// * Data structures employed during transcript operations
// *
// ------------------------------------------------------------------------------------------------------
type SmartContract struct {
	contractapi.Contract
}

// StudentInfo creates a data structure that corresponds to a relation in the relational data model of relational database management system (RDBMS)
type StudentInfo struct {
	Faculty          string `json:"faculty"`
	Department       string `json:"department"`
	StudentID        int    `json:"student_id"`
	StudentSurname   string `json:"student_surname"`
	StudentName      string `json:"student_name"`
	NationalID       string `json:"national_id"`
	RegistrationDate string `json:"registration_date"`
	RegistrationType string `json:"registration_type"`
	ProgramType      string `json:"program_type"`
	Class            int    `json:"class"`
	StudentSemester  int    `json:"student_semester"`
	HashValue        string `json:"hash_value"`
}

// TakenCourse creates a data structure that corresponds to a relation in the relational data model of relational database management system (RDBMS)
type TakenCourse struct {
	StudentID     int     `json:"student_id"`
	CourseCode    string  `json:"course_code"`
	Grade         string  `json:"grade"`
	Point         float32 `json:"point"`
	TakenSemester int     `json:"taken_semester"`
	HashValue     string  `json:"hash_value"`
}

// CourseInfo creates a data structure that corresponds to a relation in the relational data model of relational database management system (RDBMS)
type CourseInfo struct {
	CourseCode string `json:"course_code"`
	CourseName string `json:"course_name"`
	CourseType string `json:"course_type"`
	ECTS       int    `json:"ects"`
	Credit     int    `json:"credit"`
	HashValue  string `json:"hash_value"`
}

// For each created StudentInfo, TakenCourse, and CourseInfo record, a MetaInfo record is created
type MetaInfo struct {
	Owner     string `json:"owner"`      // HEI Name
	StudentID string `json:"student_id"` // Student ID
	Relation  string `json:"relation"`   // Corresponds to a relation name in RDMS
	HashValue string `json:"hash_value"` // Calculated hash value of except HashCode field
}

// Taken courses (TakenCourse) and courses info (CourseInfo) are combined to construct a transcript
type CombinedCourseRecords struct {
	CourseCode    string  `json:"course_code"`
	CourseName    string  `json:"course_name"`
	CourseType    string  `json:"course_type"`
	ECTS          int     `json:"ects"`
	Credit        int     `json:"credit"`
	Grade         string  `json:"grade"`
	Point         float32 `json:"point"`
	TakenSemester int     `json:"taken_semester"`
}

// This is the ultimate data structure that consists of StudentInfo, CourseInfo, and TakenCourses to respond to a student’s queried transcript.
type StudentTranscript struct {
	InfoStudent StudentInfo             `json:"student_informations"`
	Courses     []CombinedCourseRecords `json:"taken_courses"`
}

//------------------------------------------------------------------------------------------------------
// *
// * Basic transcript operations
// *
//------------------------------------------------------------------------------------------------------
//

func (Transcript *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	var compositeKey, generatedHashValue string

	// 1- Create studentinfos and add them to the ledger
	Students := []StudentInfo{{Faculty: "Faculty of Engineering and Architecture",
		Department:       "Department of Computer Engineering",
		StudentID:        190908809,
		StudentSurname:   "Selvi",
		StudentName:      "Osman",
		NationalID:       "44262495576",
		RegistrationDate: "02.09.2022",
		RegistrationType: "Major / OSYM",
		ProgramType:      "Undergraduate",
		Class:            1,
		StudentSemester:  1,
		HashValue:        ""},
	}

	MetaStudents := []MetaInfo{{Owner: "Fenerbahce University",
		StudentID: "190908809",
		Relation:  "StudentInfo",
		HashValue: "NULL"},
	}

	for index := range Students {
		generatedHashValue = StructToMD5(Students[index])
		Students[index].HashValue = generatedHashValue
		MetaStudents[index].HashValue = generatedHashValue

		infoJSON, err := json.Marshal(Students[index])
		if err != nil {
			return fmt.Errorf("failed to convert struct to json object: %v", err)
		}

		err = ctx.GetStub().PutState(Students[index].HashValue, infoJSON)
		if err != nil {
			return fmt.Errorf("failed to put student info to world state. %v", err)
		}

		compositeKey, err = ctx.GetStub().CreateCompositeKey("heiID", []string{MetaStudents[index].Owner, MetaStudents[index].StudentID, MetaStudents[index].HashValue})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		infoJSON, err = json.Marshal(MetaStudents[index])
		if err != nil {
			return fmt.Errorf("failed to convert struct to json object: %v", err)
		}

		err = ctx.GetStub().PutState(compositeKey, infoJSON)
		if err != nil {
			return fmt.Errorf("failed to put meta student info to world state. %v", err)
		}
	}

	// 2- Create taken courses and add them to the ledger
	TakenCourses := []TakenCourse{{StudentID: 190908809, CourseCode: "COMP1001", Grade: "AA", Point: 20, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "COMP1003", Grade: "BA", Point: 21, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "ENG103", Grade: "BB", Point: 6, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "MATH1001", Grade: "CB", Point: 18.9, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "PHYS1001", Grade: "CC", Point: 8, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "PHYS1011", Grade: "CC", Point: 4, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "TURK103", Grade: "BB", Point: 6, TakenSemester: 1, HashValue: ""},
		{StudentID: 190908809, CourseCode: "UNI103", Grade: "AA", Point: 8, TakenSemester: 1, HashValue: ""},
	}

	MetaTakenCourses := []MetaInfo{{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "TakenCourse", HashValue: "NULL"},
	}

	for index2 := range TakenCourses {
		generatedHashValue = StructToMD5(TakenCourses[index2])
		TakenCourses[index2].HashValue = generatedHashValue
		MetaTakenCourses[index2].HashValue = generatedHashValue

		infoJSON, err := json.Marshal(TakenCourses[index2])
		if err != nil {
			return fmt.Errorf("failed to convert struct to json object: %v", err)
		}

		err = ctx.GetStub().PutState(TakenCourses[index2].HashValue, infoJSON)
		if err != nil {
			return fmt.Errorf("failed to put course record to world state. %v", err)
		}

		compositeKey, err = ctx.GetStub().CreateCompositeKey("heiID", []string{MetaTakenCourses[index2].Owner, MetaTakenCourses[index2].StudentID, MetaTakenCourses[index2].HashValue})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		infoJSON, err = json.Marshal(MetaTakenCourses[index2])
		if err != nil {
			return fmt.Errorf("failed to convert struct to json object: %v", err)
		}

		err = ctx.GetStub().PutState(compositeKey, infoJSON)
		if err != nil {
			return fmt.Errorf("failed to put meta student info to world state. %v", err)
		}

	}

	// 3- Create course infos and add them to the ledger
	CourseInfoS := []CourseInfo{{CourseCode: "COMP1001", CourseName: "Fundamentals of Computer Engineering", CourseType: "C", ECTS: 5, Credit: 3, HashValue: ""},
		{CourseCode: "COMP1003", CourseName: "Algorithms and Programming I", CourseType: "C", ECTS: 6, Credit: 3, HashValue: ""},
		{CourseCode: "ENG103", CourseName: "Advanced English I", CourseType: "C", ECTS: 2, Credit: 2, HashValue: ""},
		{CourseCode: "MATH1001", CourseName: "Calculus I", CourseType: "C", ECTS: 7, Credit: 4, HashValue: ""},
		{CourseCode: "PHYS1001", CourseName: "Physics I", CourseType: "C", ECTS: 4, Credit: 3, HashValue: ""},
		{CourseCode: "PHYS1011", CourseName: "Physics I Laboratory", CourseType: "C", ECTS: 2, Credit: 1, HashValue: ""},
		{CourseCode: "TURK103", CourseName: "Turkish Language I", CourseType: "C", ECTS: 2, Credit: 2, HashValue: ""},
		{CourseCode: "UNI103", CourseName: "University Life and Culture", CourseType: "C", ECTS: 2, Credit: 2, HashValue: ""},
	}

	MetaCourseInfoS := []MetaInfo{{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
		{Owner: "Fenerbahce University", StudentID: "190908809", Relation: "CourseInfo", HashValue: "NULL"},
	}

	for index3 := range CourseInfoS {
		generatedHashValue = StructToMD5(CourseInfoS[index3])
		CourseInfoS[index3].HashValue = generatedHashValue
		MetaCourseInfoS[index3].HashValue = generatedHashValue

		infoJSON, err := json.Marshal(CourseInfoS[index3])
		if err != nil {
			return fmt.Errorf("failed to convert struct to json object: %v", err)
		}

		err = ctx.GetStub().PutState(CourseInfoS[index3].HashValue, infoJSON)
		if err != nil {
			return fmt.Errorf("failed to put course info record to world state. %v", err)
		}

		compositeKey, err = ctx.GetStub().CreateCompositeKey("heiID", []string{MetaCourseInfoS[index3].Owner, MetaCourseInfoS[index3].StudentID, MetaCourseInfoS[index3].HashValue})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		infoJSON, err = json.Marshal(MetaCourseInfoS[index3])
		if err != nil {
			return fmt.Errorf("failed to convert struct to json object: %v", err)
		}

		err = ctx.GetStub().PutState(compositeKey, infoJSON)
		if err != nil {
			return fmt.Errorf("failed to put meta course info to world state. %v", err)
		}

	}

	return nil
}

func StructToMD5(incomingStruct interface{}) (generatedHashValue string) {
	generatedString := StructToString(incomingStruct)
	generatedHashValue = StringToMD5(generatedString)
	return
}

func StructToString(incomingStruct interface{}) (result string) {
	values := reflect.ValueOf(incomingStruct)
	numberOfFields := values.NumField()
	mySlice := make([]string, numberOfFields)

	for i := 0; i < numberOfFields; i++ {
		fieldType := values.Type().Field(i).Name
		fmt.Println(fieldType)
		if fieldType != "HashCode" { // Convert everthing to a string with comma expect from HashCode field
			value := reflect.ValueOf(values.Field(i)).Interface()
			stringData := fmt.Sprintf("%v", value)
			mySlice[i] = stringData
		}
	}

	result = strings.Join(mySlice, ",") // This line is inserting comma after each element. We don't wanna to insert after the last element a comma. Because of that we need to delete last comma
	result = result[:len(result)-1]
	return
}

func StringToMD5(value string) string {
	new_hasher := md5.New()
	new_hasher.Write([]byte(value))
	return hex.EncodeToString(new_hasher.Sum(nil))
}

func (Transcript *SmartContract) IsRecordExists(ctx contractapi.TransactionContextInterface, Owner string, StudentID string, HashCode string) (bool, error) {
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s","student_id":"%s", "hash_value":"%s"}}`, Owner, StudentID, HashCode)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)

	if err != nil {
		return true, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if resultsIterator.HasNext() {
		return true, nil
	}

	defer resultsIterator.Close()

	return false, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * To create and include new records to Hyperledger Fabric
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) InsertNewRecordStudentInfo(ctx contractapi.TransactionContextInterface, owner string, faculty string, department string,
	studentId int, surname string, name string, nationalid string, registrationdate string, registrationtype string, programtype string, class int, semester int) (bool, error) {

	var err error
	var compositeKey, generatedHashValue string
	var IsExist bool
	var student StudentInfo
	var meta MetaInfo

	student.Faculty = faculty
	student.Department = department
	student.StudentID = studentId
	student.StudentSurname = surname
	student.StudentName = name
	student.NationalID = nationalid
	student.RegistrationDate = registrationdate
	student.RegistrationType = registrationtype
	student.ProgramType = programtype
	student.Class = class
	student.StudentSemester = semester

	generatedHashValue = StructToMD5(student)
	student.HashValue = generatedHashValue

	IsExist, err = Transcript.IsRecordExists(ctx, owner, strconv.Itoa(studentId), generatedHashValue)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	if IsExist {
		return false, fmt.Errorf("the record you sent exists: %v", err)
	}

	jsonStudent, err := json.Marshal(student)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to json object: %v", err)
	}

	err = ctx.GetStub().PutState(student.HashValue, jsonStudent)
	if err != nil {
		return false, fmt.Errorf("failed to put student info to world state. %v", err)
	}

	meta.Owner = owner
	meta.StudentID = strconv.Itoa(studentId)
	meta.Relation = "StudentInfo"
	meta.HashValue = generatedHashValue

	compositeKey, err = ctx.GetStub().CreateCompositeKey("heiID", []string{meta.Owner, meta.StudentID, meta.HashValue})
	if err != nil {
		return false, fmt.Errorf("failed to create composite key: %v", err)
	}

	jsonMeta, err := json.Marshal(meta)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to json object: %v", err)
	}

	err = ctx.GetStub().PutState(compositeKey, jsonMeta)
	if err != nil {
		return false, fmt.Errorf("failed to put meta student info to world state. %v", err)
	}

	return true, nil
}

func (Transcript *SmartContract) InsertNewRecordTakenCourse(ctx contractapi.TransactionContextInterface, owner string, studentId int,
	courseCode string, grade string, point float32, takenSemester int) (bool, error) {

	var err error
	var compositeKey, generatedHashValue string
	var IsExist bool
	var course TakenCourse
	var meta MetaInfo

	course.StudentID = studentId
	course.CourseCode = courseCode
	course.Grade = grade
	course.Point = point
	course.TakenSemester = takenSemester

	generatedHashValue = StructToMD5(course)
	course.HashValue = generatedHashValue

	IsExist, err = Transcript.IsRecordExists(ctx, owner, strconv.Itoa(studentId), generatedHashValue)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	if IsExist {
		return false, fmt.Errorf("the record you sent exists: %v", err)
	}

	jsonCourse, err := json.Marshal(course)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to json object: %v", err)
	}

	err = ctx.GetStub().PutState(course.HashValue, jsonCourse)
	if err != nil {
		return false, fmt.Errorf("failed to put student info to world state. %v", err)
	}

	meta.Owner = owner
	meta.StudentID = strconv.Itoa(studentId)
	meta.Relation = "TakenCourse"
	meta.HashValue = generatedHashValue

	compositeKey, err = ctx.GetStub().CreateCompositeKey("heiID", []string{meta.Owner, meta.StudentID, meta.HashValue})
	if err != nil {
		return false, fmt.Errorf("failed to create composite key: %v", err)
	}

	jsonMeta, err := json.Marshal(meta)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to json object: %v", err)
	}

	err = ctx.GetStub().PutState(compositeKey, jsonMeta)
	if err != nil {
		return false, fmt.Errorf("failed to put meta student info to world state. %v", err)
	}

	return true, nil
}

func (Transcript *SmartContract) InsertNewRecordCourseInfo(ctx contractapi.TransactionContextInterface, owner string, studentnumber int,
	courseCode string, courseName string, courseType string, ects int, credit int) (bool, error) {

	var err error
	var compositeKey, generatedHashValue string
	var IsExist bool
	var InfoCourse CourseInfo
	var meta MetaInfo

	InfoCourse.CourseCode = courseCode
	InfoCourse.CourseName = courseName
	InfoCourse.CourseType = courseType
	InfoCourse.Credit = credit
	InfoCourse.ECTS = ects

	generatedHashValue = StructToMD5(InfoCourse)
	InfoCourse.HashValue = generatedHashValue

	IsExist, err = Transcript.IsRecordExists(ctx, owner, strconv.Itoa(studentnumber), generatedHashValue)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	if IsExist {
		return false, fmt.Errorf("the record you sent exists: %v", err)
	}

	jsonCourse, err := json.Marshal(InfoCourse)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to json object: %v", err)
	}

	err = ctx.GetStub().PutState(InfoCourse.HashValue, jsonCourse)
	if err != nil {
		return false, fmt.Errorf("failed to put student info to world state. %v", err)
	}

	meta.Owner = owner
	meta.StudentID = strconv.Itoa(studentnumber)
	meta.Relation = "CourseInfo"
	meta.HashValue = generatedHashValue

	compositeKey, err = ctx.GetStub().CreateCompositeKey("heiID", []string{meta.Owner, meta.StudentID, meta.HashValue})
	if err != nil {
		return false, fmt.Errorf("failed to create composite key: %v", err)
	}

	jsonMeta, err := json.Marshal(meta)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to json object: %v", err)
	}

	err = ctx.GetStub().PutState(compositeKey, jsonMeta)
	if err != nil {
		return false, fmt.Errorf("failed to put meta student info to world state. %v", err)
	}
	return true, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Get a student's student info: it is a relation of a relational data model
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) Get_Student_StudentInfo(ctx contractapi.TransactionContextInterface, hei string, studentID string) (*StudentInfo, error) {
	var recordStudentInfo *StudentInfo
	var err error
	var hashValueofStudentInfo []string

	hashValueofStudentInfo, err = Transcript.Get_Student_StudentInfo_HashValues(ctx, hei, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	for index := 0; index < len(hashValueofStudentInfo); index++ {
		infoStudent, err := Transcript.Get_StudentInfo_ByHashValue(ctx, hashValueofStudentInfo[index])
		if err != nil {
			return nil, fmt.Errorf("error during fetch taken course record by hash value: %v", err)
		}

		recordStudentInfo = infoStudent
	}

	return recordStudentInfo, nil
}

func (Transcript *SmartContract) Get_Student_StudentInfo_HashValues(ctx contractapi.TransactionContextInterface, hei string, studentID string) ([]string, error) {
	var iterator shim.StateQueryIteratorInterface
	var err error

	var hashValues []string

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s", "relation":"StudentInfo", "student_id":"%s"}}`, hei, studentID)

	iterator, err = ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if !iterator.HasNext() {
		return nil, fmt.Errorf("no record were found relevant to the given arguments on worldstate db")
	}

	defer iterator.Close()

	for iterator.HasNext() {
		var record MetaInfo
		queryRow, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over the returned records : %v", err)
		}
		err = json.Unmarshal(queryRow.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
		}

		hashValues = append(hashValues, record.HashValue)
	}

	return hashValues, nil
}

func (Transcript *SmartContract) Get_StudentInfo_ByHashValue(ctx contractapi.TransactionContextInterface, hashValue string) (*StudentInfo, error) {
	jsonData, err := ctx.GetStub().GetState(hashValue)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if jsonData == nil {
		return nil, fmt.Errorf("there is not a record with the given hash value: %v", hashValue)
	}

	var infoStudent StudentInfo
	err = json.Unmarshal(jsonData, &infoStudent)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
	}

	return &infoStudent, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Get a student's course infos: it is a relation of a relational data model
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) Get_Student_CourseInfos(ctx contractapi.TransactionContextInterface, hei string, studentID string) ([]*CourseInfo, error) {
	var recordsCourseInfos []*CourseInfo
	var err error
	var hashValuesofCourseInfos []string

	hashValuesofCourseInfos, err = Transcript.Get_Student_CourseInfos_HashValues(ctx, hei, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	for index := 0; index < len(hashValuesofCourseInfos); index++ {
		course, err := Transcript.Get_CourseInfo_ByHashValue(ctx, hashValuesofCourseInfos[index])
		if err != nil {
			return nil, fmt.Errorf("error during fetch taken course record by hash value: %v", err)
		}

		recordsCourseInfos = append(recordsCourseInfos, course)
	}

	return recordsCourseInfos, nil
}

func (Transcript *SmartContract) Get_Student_CourseInfos_HashValues(ctx contractapi.TransactionContextInterface, hei string, studentID string) ([]string, error) {
	var iterator shim.StateQueryIteratorInterface
	var err error

	var hashValues []string

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s", "relation":"CourseInfo", "student_id":"%s"}}`, hei, studentID)

	iterator, err = ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if !iterator.HasNext() {
		return nil, fmt.Errorf("no record were found relevant to the given arguments on worldstate db")
	}

	defer iterator.Close()

	for iterator.HasNext() {
		var record MetaInfo
		queryRow, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over the returned records : %v", err)
		}
		err = json.Unmarshal(queryRow.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
		}

		hashValues = append(hashValues, record.HashValue)
	}

	return hashValues, nil
}

func (Transcript *SmartContract) Get_CourseInfo_ByHashValue(ctx contractapi.TransactionContextInterface, hashValue string) (*CourseInfo, error) {
	jsonData, err := ctx.GetStub().GetState(hashValue)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if jsonData == nil {
		return nil, fmt.Errorf("there is not a record with the given hash value: %v", hashValue)
	}

	var infoCourse CourseInfo
	err = json.Unmarshal(jsonData, &infoCourse)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
	}

	return &infoCourse, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Get a student's taken courses: it is a relation of a relational data model
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) Get_Student_TakenCourses(ctx contractapi.TransactionContextInterface, hei string, studentID string) ([]*TakenCourse, error) {
	var recordsTakenCourses []*TakenCourse
	var err error
	var hashValuesofTakenCourses []string

	hashValuesofTakenCourses, err = Transcript.Get_Student_TakenCourses_HashValues(ctx, hei, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	for index := 0; index < len(hashValuesofTakenCourses); index++ {
		course, err := Transcript.Get_TakenCourse_ByHashValue(ctx, hashValuesofTakenCourses[index])
		if err != nil {
			return nil, fmt.Errorf("error during fetch taken course record by hash value: %v", err)
		}

		// Get course infoyu da alıp birleştirmeli burada

		recordsTakenCourses = append(recordsTakenCourses, course)
	}

	return recordsTakenCourses, nil
}

func (Transcript *SmartContract) Get_Student_TakenCourses_HashValues(ctx contractapi.TransactionContextInterface, hei string, studentID string) ([]string, error) {
	var iterator shim.StateQueryIteratorInterface
	var err error

	var hashValues []string

	queryString := fmt.Sprintf(`{"selector":{"owner":"%s", "relation":"TakenCourse", "student_id":"%s"}}`, hei, studentID)

	iterator, err = ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if !iterator.HasNext() {
		return nil, fmt.Errorf("no record were found relevant to the given arguments on worldstate db")
	}

	defer iterator.Close()

	for iterator.HasNext() {
		var record MetaInfo
		queryRow, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over the returned records : %v", err)
		}
		err = json.Unmarshal(queryRow.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
		}

		hashValues = append(hashValues, record.HashValue)
	}

	return hashValues, nil
}

func (Transcript *SmartContract) Get_TakenCourse_ByHashValue(ctx contractapi.TransactionContextInterface, hashValue string) (*TakenCourse, error) {
	jsonData, err := ctx.GetStub().GetState(hashValue)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if jsonData == nil {
		return nil, fmt.Errorf("there is not a record with the given hash value: %v", hashValue)
	}

	var takenCourse TakenCourse
	err = json.Unmarshal(jsonData, &takenCourse)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
	}

	return &takenCourse, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Get a higher education institution's (HEI's) taken courses by students
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) Get_HEI_TakenCourses(ctx contractapi.TransactionContextInterface, hei string) ([]*TakenCourse, error) {
	var records []*MetaInfo
	var recordsStudentInfo []*TakenCourse
	var err error

	records, err = Transcript.Get_HEI_MetaInfos_TakenCourses(ctx, hei)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	for index := 0; index < len(records); index++ {
		recordStudentInfo, err := Transcript.Get_TakenCourse_ByHashValue(ctx, records[index].HashValue)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		recordsStudentInfo = append(recordsStudentInfo, recordStudentInfo)
	}
	return recordsStudentInfo, nil
}

func (Transcript *SmartContract) Get_HEI_MetaInfos_TakenCourses(ctx contractapi.TransactionContextInterface, hei string) ([]*MetaInfo, error) {
	var iterator shim.StateQueryIteratorInterface
	var records []*MetaInfo
	var err error

	var queryString = fmt.Sprintf(`{"selector":{"owner":"%s", "relation":"TakenCourse"}}`, hei)
	iterator, err = ctx.GetStub().GetQueryResult(queryString)

	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if !iterator.HasNext() {
		return nil, fmt.Errorf("no record were found relevant to the given arguments on worldstate db")
	}

	defer iterator.Close()

	for iterator.HasNext() {
		queryRow, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over the returned records : %v", err)
		}
		var record MetaInfo
		err = json.Unmarshal(queryRow.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
		}

		records = append(records, &record)
	}

	return records, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Get a higher education institution's (HEI's) students info by students
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) Get_HEI_StudentInfos(ctx contractapi.TransactionContextInterface, hei string) ([]*StudentInfo, error) {
	var records []*MetaInfo
	var recordsStudentInfo []*StudentInfo
	var err error

	records, err = Transcript.Get_HEI_MetaInfos_StudentInfos(ctx, hei)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	for index := 0; index < len(records); index++ {
		recordStudentInfo, err := Transcript.Get_StudentInfo_ByHashValue(ctx, records[index].HashValue)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		recordsStudentInfo = append(recordsStudentInfo, recordStudentInfo)
	}
	return recordsStudentInfo, nil
}

func (Transcript *SmartContract) Get_HEI_MetaInfos_StudentInfos(ctx contractapi.TransactionContextInterface, hei string) ([]*MetaInfo, error) {
	var iterator shim.StateQueryIteratorInterface
	var records []*MetaInfo
	var err error

	var queryString = fmt.Sprintf(`{"selector":{"owner":"%s", "relation":"StudentInfo"}}`, hei)
	iterator, err = ctx.GetStub().GetQueryResult(queryString)

	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if !iterator.HasNext() {
		return nil, fmt.Errorf("no record were found relevant to the given arguments on worldstate db")
	}

	defer iterator.Close()

	for iterator.HasNext() {
		queryRow, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over the returned records : %v", err)
		}
		var record MetaInfo
		err = json.Unmarshal(queryRow.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
		}

		records = append(records, &record)
	}

	return records, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Get a higher education institution's (HEI's) course info by students
// *
//------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) Get_HEI_CourseInfos(ctx contractapi.TransactionContextInterface, hei string) ([]*CourseInfo, error) {
	var records []*MetaInfo
	var recordsStudentInfo []*CourseInfo
	var err error

	records, err = Transcript.Get_HEI_MetaInfos_CourseInfos(ctx, hei)
	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	for index := 0; index < len(records); index++ {
		recordStudentInfo, err := Transcript.Get_CourseInfo_ByHashValue(ctx, records[index].HashValue)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		recordsStudentInfo = append(recordsStudentInfo, recordStudentInfo)
	}
	return recordsStudentInfo, nil
}

func (Transcript *SmartContract) Get_HEI_MetaInfos_CourseInfos(ctx contractapi.TransactionContextInterface, hei string) ([]*MetaInfo, error) {
	var iterator shim.StateQueryIteratorInterface
	var records []*MetaInfo
	var err error

	var queryString = fmt.Sprintf(`{"selector":{"owner":"%s", "relation":"CourseInfo"}}`, hei)
	iterator, err = ctx.GetStub().GetQueryResult(queryString)

	if err != nil {
		return nil, fmt.Errorf("failed to read from worldstate db : %v", err)
	}

	if !iterator.HasNext() {
		return nil, fmt.Errorf("no record were found relevant to the given arguments on worldstate db")
	}

	defer iterator.Close()

	for iterator.HasNext() {
		queryRow, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over the returned records : %v", err)
		}
		var record MetaInfo
		err = json.Unmarshal(queryRow.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch json data to struct : %v", err)
		}

		records = append(records, &record)
	}

	return records, nil
}

//------------------------------------------------------------------------------------------------------
// *
// * Construct a student's transcript
// *
// ------------------------------------------------------------------------------------------------------

func (Transcript *SmartContract) GetStudentTranscript(ctx contractapi.TransactionContextInterface, hei string, studentID string) (*StudentTranscript, error) {
	var new_transcript StudentTranscript
	var coursesTakenbyStudent []CombinedCourseRecords

	var infoStudent *StudentInfo
	var infoCourses []*CourseInfo
	var coursesTaken []*TakenCourse

	var err error

	infoStudent, err = Transcript.Get_Student_StudentInfo(ctx, hei, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to construct the transcript from the world state db: %v", err)
	}

	infoCourses, err = Transcript.Get_Student_CourseInfos(ctx, hei, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to construct the transcript from the world state db: %v", err)
	}

	coursesTaken, err = Transcript.Get_Student_TakenCourses(ctx, hei, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to construct the transcript from the world state db: %v", err)
	}

	for _, course := range coursesTaken {

		var newCourseCombined CombinedCourseRecords
		newCourseCombined.CourseCode = course.CourseCode
		newCourseCombined.Grade = course.Grade
		newCourseCombined.Point = course.Point
		newCourseCombined.TakenSemester = course.TakenSemester

		for _, info := range infoCourses {
			if course.CourseCode == info.CourseCode {
				newCourseCombined.CourseName = info.CourseName
				newCourseCombined.CourseType = info.CourseType
				newCourseCombined.ECTS = info.ECTS
				newCourseCombined.Credit = info.Credit
				coursesTakenbyStudent = append(coursesTakenbyStudent, newCourseCombined)
			}
		}

	}

	new_transcript.InfoStudent = *infoStudent
	new_transcript.Courses = coursesTakenbyStudent

	return &new_transcript, nil

}

//------------------------------------------------------------------------------------------------------
// *
// *												THE END
// *
// ------------------------------------------------------------------------------------------------------
