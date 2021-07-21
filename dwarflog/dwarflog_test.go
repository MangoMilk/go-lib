package dwarflog

import "testing"

func setup() {
	Setup(&Config{
		Format:     JsonFormat,
		Rolling:    true,
		Path:       "./log",
		FilePrefix: "dwaftlog",
	})
}

func teardown() {

}

func TestMain(m *testing.M) {
	setup()
	m.Run()
	teardown()
}

func TestInfoLog(t *testing.T) {

	// log text
	Info("123")

	// log map
	m := map[string]interface{}{"a": 1, "b": "b"}
	Info(m)

	// log map and slice
	s := []interface{}{"123", 12}
	Info(m, s)

	// log struct
	type Res struct {
		Code    int
		Message string
		Content interface{}
	}
	type SuccessRes struct {
		IsSuccess int
	}

	r := Res{
		Code:    0,
		Message: "success",
		Content: SuccessRes{IsSuccess: 1},
	}
	Info("【Response】", r)

	//Infof()
	//Infoln()
}

func TestErrorLog(t *testing.T) {
	Error("123")

	m := map[string]interface{}{"a": 1, "b": "b"}
	Error(m)

	s := []interface{}{"123", 12}
	Error(m, s)

	type Res struct {
		Code    int
		Message string
		Content interface{}
	}
	type SuccessRes struct {
		IsSuccess int
	}

	r := Res{
		Code:    0,
		Message: "success",
		Content: SuccessRes{IsSuccess: 1},
	}
	Error("【Response】", r)
}
