package pystate

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"testing"
)

func TestCreateState(t *testing.T) {
	cc := &core.ContextConfig{}
	ctx := core.NewContext(cc)
	Convey("Given a pystate creator", t, func() {
		ct := Creator{}
		Convey("When the parameter has all required values", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
				"class_name":  data.String("TestClass"),
			}
			st, err := ct.CreateState(ctx, params)
			So(err, ShouldBeNil)
			Reset(func() {
				st.Terminate(ctx)
			})
			ps, ok := st.(*state)
			So(ok, ShouldBeTrue)

			Convey("Then the state should be created and set default value", func() {
				So(ps.base.params.ModulePath, ShouldEqual, "")
				So(ps.base.params.ModuleName, ShouldEqual, "_test_creator_module")

				ctx.SharedStates.Add("creator_test", "creator_test", st)
				Reset(func() {
					ctx.SharedStates.Remove("creator_test")
				})
				Convey("When prepare to be called one instance method", func() {
					Convey("Then exist instance method should be called", func() {
						dt := data.String("test")
						v, err := CallMethod(ctx, "creator_test", "write", dt)
						So(err, ShouldBeNil)
						So(v, ShouldEqual, `called! arg is "test"`)
					})
					Convey("Then not exist instance method should not be called and return error", func() {
						_, err = CallMethod(ctx, "creator_test", "not_exist_method")
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("Then calling Terminate actually terminates the state", func() {
				So(ps.base.CheckTermination(), ShouldBeNil)
				So(ps.Terminate(ctx), ShouldBeNil)
				So(ps.base.CheckTermination(), ShouldPointTo, ErrAlreadyTerminated)
			})
		})

		Convey("When the parameter has constructor arguments", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
				"class_name":  data.String("TestClass2"),
				"v1":          data.String("init_test"),
				"v2":          data.String("init_test2"),
			}
			Convey("Then the state should be created with constructor arguments", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})

				ctx.SharedStates.Add("creator_test2", "creator_test2", state)
				Reset(func() {
					ctx.SharedStates.Remove("creator_test2")
				})
				v, err := CallMethod(ctx, "creator_test2", "confirm")
				So(err, ShouldBeNil)
				So(v, ShouldEqual, `constructor init arg is v1=init_test, v2=init_test2`)
			})
		})

		Convey("When the parameter has constructor arguments which lack optional value", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
				"class_name":  data.String("TestClass3"),
				"a":           data.Int(55),
			}
			Convey("Then the state should be created and initialized with only required value", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})

				ctx.SharedStates.Add("creator_test3", "creator_test3", state)
				Reset(func() {
					ctx.SharedStates.Remove("creator_test3")
				})
				v, err := CallMethod(ctx, "creator_test3", "confirm")
				So(err, ShouldBeNil)
				So(v, ShouldEqual, "constructor init arg is a=55, b=b, c={}")
			})
		})

		Convey("When the parameter has required value and option value", func() {
			params := data.Map{
				"module_name":  data.String("_test_creator_module"),
				"class_name":   data.String("TestClass"),
				"write_method": data.String("write"),
			}
			Convey("Then the state should be created and it is writable", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})
				ps, ok := state.(*writableState)
				So(ok, ShouldBeTrue)
				So(ps.base.params.WriteMethodName, ShouldEqual, "write")

				t := &core.Tuple{}
				err = ps.Write(ctx, t)
				So(err, ShouldBeNil)
			})
		})

		Convey("When the parameter lacks module name", func() {
			params := data.Map{
				"class_name": data.String("TestClass"),
			}
			Convey("Then a state should not be created", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "module_name")
				So(state, ShouldBeNil)
			})
		})

		Convey("When the parameter lacks class name", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
			}
			Convey("Then a state should not be created", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "class_name")
				So(state, ShouldBeNil)
			})
		})
	})
}

func TestSaveLoadState(t *testing.T) {
	params := data.Map{
		"a": data.Int(1),
		"b": data.String("hoge"),
	}

	Convey("Given a saved state", t, func() {
		ctx := core.NewContext(nil)
		c := Creator{}
		state, err := c.CreateState(ctx, data.Map{
			"module_name": data.String("_test_creator_module"),
			"class_name":  data.String("TestClass4"),
			"a":           params["a"],
			"b":           params["b"],
		})
		So(err, ShouldBeNil)
		Reset(func() {
			state.Terminate(ctx)
		})
		So(ctx.SharedStates.Add("creator_test4", "py", state), ShouldBeNil)
		s := state.(core.LoadableSharedState)
		buf := bytes.NewBuffer(nil)
		So(s.Save(ctx, buf, params), ShouldBeNil)

		Convey("When modifying parameters", func() {
			_, err := CallMethod(ctx, "creator_test4", "modify_params")
			So(err, ShouldBeNil)

			Convey("Then they should be different from the originals", func() {
				p, err := CallMethod(ctx, "creator_test4", "confirm")
				So(err, ShouldBeNil)
				So(p, ShouldNotResemble, params)
			})
		})

		Convey("When loading the state", func() {
			_, err := CallMethod(ctx, "creator_test4", "modify_params")
			So(err, ShouldBeNil)
			So(s.Load(ctx, buf, data.Map{}), ShouldBeNil)

			Convey("Then it should preserve the original parameters", func() {
				p, err := CallMethod(ctx, "creator_test4", "confirm")
				So(err, ShouldBeNil)
				So(p, ShouldResemble, params)
			})
		})

		Convey("When loading the state as a new one", func() {
			// This parameter modification after Save shouldn't affect the
			// newly loaded state.
			_, err := CallMethod(ctx, "creator_test4", "modify_params")
			So(err, ShouldBeNil)

			s2, err := c.LoadState(ctx, buf, data.Map{})
			So(err, ShouldBeNil)
			Reset(func() {
				s2.Terminate(ctx)
			})
			So(ctx.SharedStates.Add("creator_test4_2", "py", s2), ShouldBeNil)

			Convey("Then it should have the same parameter as the original", func() {
				p, err := CallMethod(ctx, "creator_test4_2", "confirm")
				So(err, ShouldBeNil)
				So(p, ShouldResemble, params)
			})
		})
	})
}

func TestPyStateTerminate(t *testing.T) {
	ctx := core.NewContext(nil)
	Convey("Given a state set python instance", t, func() {
		c := Creator{}
		s, err := c.CreateState(ctx, data.Map{
			"module_name": data.String("_test_creator_module"),
			"class_name":  data.String("TestClassTerminateError"),
		})
		So(err, ShouldBeNil)
		Convey("When call terminate", func() {
			err := s.Terminate(ctx)
			Convey("Then error should be occurred from python", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "ZeroDivisionError")
			})
		})
	})
}
