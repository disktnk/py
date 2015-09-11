package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ugorji/go/codec"
	"pfi/sensorbee/sensorbee/data"
	"testing"
)

func init() {
	// goconvey is call same function several times, so in order to call
	// `Initialize` only once, use `init`
	Initialize()
}

func TestPythonCall(t *testing.T) {
	Convey("Given an initialized python interpreter", t, func() {

		ImportSysAndAppendPath("")

		Convey("When get invalid module", func() {
			_, err := LoadModule("notexist")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
			})
		})
		Convey("When get valid module", func() {
			mdl, err := LoadModule("_test")
			Convey("Then process should get PyModule", func() {
				So(err, ShouldBeNil)
				So(mdl, ShouldNotBeNil)
				Reset(func() {
					mdl.DecRef()
				})

				Convey("And when call invalid function", func() {
					_, err := mdl.CallIntInt("notFoundMethod", 1)
					Convey("Then an error should be occurred", func() {
						So(err, ShouldNotBeNil)
					})
				})

				Convey("And when call int-int function", func() {
					actual, err := mdl.CallIntInt("tenTimes", 3)
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, 30)
					})
				})

				Convey("And when call int-int function with Call method", func() {
					actual, err := mdl.Call("tenTimes", data.Int(4))
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, data.Int(40))
					})
				})

				Convey("And when call none-string function", func() {
					actual, err := mdl.CallNoneString("logger")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "called")
					})
				})

				Convey("And when call none-string function with Call", func() {
					actual, err := mdl.Call("logger")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, data.String("called"))
					})
				})

				Convey("And when call none-2string function", func() {
					actual1, actual2, err := mdl.CallNone2String("twoLogger")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual1, ShouldEqual, "called1")
						So(actual2, ShouldEqual, "called2")
					})
				})

				Convey("And when call string-string function", func() {
					actual, err := mdl.CallStringString("plusSuffix", "test")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "test_through_python")
					})
				})

				Convey("And when call specialized function", func() {
					h := &codec.MsgpackHandle{}
					var out []byte
					enc := codec.NewEncoderBytes(&out, h)

					dat := []float32{1.1, 1.2, 1.3, 1.4, 1.5}
					data2 := []int{9, 8, 7., 6, 5}
					ds := map[string]interface{}{}
					ds["data"] = dat
					ds["target"] = data2
					ds["model"] = []byte{}
					err := enc.Encode(ds)
					So(err, ShouldBeNil)
					So(len(out), ShouldNotEqual, 0)
					actual1, err := mdl.CallByteByte("loadMsgPack", out)
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(len(actual1), ShouldNotEqual, 0)

						var ds2 map[string]interface{}
						dec := codec.NewDecoderBytes(actual1, h)
						dec.Decode(&ds2)
						model, ok := ds2["model"]
						So(ok, ShouldBeTrue)
						modelByte, ok := model.([]byte)
						So(ok, ShouldBeTrue)
						So(len(modelByte), ShouldNotEqual, 0)

						log, ok := ds2["log"]
						So(ok, ShouldBeTrue)
						logByte, ok := log.([]byte)
						So(ok, ShouldBeTrue)
						So(string(logByte), ShouldEqual, "done")

						Convey("And when pass pickle data", func() {
							var out2 []byte
							enc2 := codec.NewEncoderBytes(&out2, h)
							err = enc2.Encode(ds2)
							So(err, ShouldBeNil)

							Convey("Then function should return model again", func() {
								actual2, err := mdl.CallByteByte("loadMsgPack", out2)
								So(err, ShouldBeNil)
								So(len(actual2), ShouldNotEqual, 0)
								So(string(actual2), ShouldResemble,
									"\x82\xa5model\xae\x80\x02U\aTEST_req\x00.\xa3log\xa4done")

							})

							Convey("Then function should return model again with Call", func() {
								actual2, err := mdl.Call("loadMsgPack", data.Blob(out2))
								So(err, ShouldBeNil)
								dat, err := data.AsString(actual2)
								So(err, ShouldBeNil)
								So(string(dat), ShouldResemble,
									"\x82\xa5model\xae\x80\x02U\aTEST_req\x00.\xa3log\xa4done")
							})
						})
					})
				})

				Convey("And when call dictionary argument function", func() {
					arg := data.Map{
						"string": data.String("test"),
						"int":    data.Int(9),
						"byte":   data.Blob([]byte("ABC")),
					}

					actual, err := mdl.CallMapString("dict", arg)
					Convey("Then function should return valid values", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "test9ABC")
					})
				})

				Convey("And when call dictionary argument function with Call", func() {
					arg := data.Map{
						"string": data.String("test"),
						"int":    data.Int(8),
						"byte":   data.Blob([]byte("ABC")),
					}

					actual, err := mdl.Call("dict", arg)
					Convey("Then function should return valid values", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, data.String("test8ABC"))
					})
				})
			})
		})
	})
}

func TestPythonInstanceCall(t *testing.T) {
	Convey("Given an initialized python module", t, func() {

		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_instance_test")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		Convey("When get an invalid class instance", func() {
			_, err := mdl.GetInstance("NonexistentClass")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When get a new test python instance", func() {
			ins, err := mdl.GetInstance("PythonTest")
			Convey("Then process should get PyModule", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
				Reset(func() {
					ins.DecRef()
				})

				Convey("And when call a logger function", func() {
					actual, err := ins.CallStringString("logger", "test")
					Convey("Then process should get a string", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "initialized_test")
					})

					Convey("And when call a logger function again", func() {
						actual2, err := ins.CallStringString("logger", "test")
						Convey("Then process should get a string", func() {
							So(err, ShouldBeNil)
							So(actual2, ShouldEqual, "initialized_test_test")
						})

						Convey("And when get a new test python instance", func() {
							ins2, err := mdl.GetInstance("PythonTest")
							Convey("Then process should get PyModule", func() {
								So(err, ShouldBeNil)
								So(ins2, ShouldNotBeNil)
								Reset(func() {
									ins2.DecRef()
								})

								Convey("And when call a logger function", func() {
									actual3, err1 := ins.CallStringString("logger", "t")
									actual4, err2 := ins2.CallStringString("logger", "t")
									Convey("Then process should get a string", func() {
										So(err1, ShouldBeNil)
										So(err2, ShouldBeNil)
										So(actual3, ShouldEqual, "initialized_test_test_t")
										So(actual4, ShouldEqual, "initialized_t")
									})
								})
							})
						})
					})
				})
			})
		})

		Convey("When get a new test python instance with param", func() {
			params := data.String("python_test")
			ins, err := mdl.GetInstance("PythonTest2", params)

			Convey("Then process should get instance and set values", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
				Reset(func() {
					ins.DecRef()
				})

				actual, err := ins.Call("get_a")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "python_test")

				Convey("And when get another python instance with param", func() {
					params2 := data.String("python_test2")
					ins2, err := mdl.GetInstance("PythonTest2", params2)

					Convey("Then process should get another instance and set values", func() {
						So(err, ShouldBeNil)
						So(ins2, ShouldNotBeNil)
						Reset(func() {
							ins2.DecRef()
						})

						actual2, err := ins2.Call("get_a")
						So(err, ShouldBeNil)
						So(actual2, ShouldEqual, "python_test2")

						react, err := ins.Call("get_a")
						So(err, ShouldBeNil)
						So(react, ShouldEqual, "python_test")

					})
				})
			})
		})
	})
}
