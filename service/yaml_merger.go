package service

import (
	// Base packages.

	"fmt"
	"reflect"

	"github.com/Sirupsen/logrus"

	// Third party packages.
	"gopkg.in/yaml.v2"
)

//MergeYaml merges yml1 into yml2 and return merged result
func MergeYaml(yml1 []byte, yml2 []byte) ([]byte, error) {
	var obj1 interface{}
	var obj2 interface{}
	if err := yaml.Unmarshal(yml1, &obj1); err != nil {
		logrus.Errorf("parse yaml got error:%v, please check file format:\n%s", err, string(yml1))
		return nil, err
	}
	if err := yaml.Unmarshal(yml2, &obj2); err != nil {
		logrus.Errorf("parse yaml got error:%v, please check file format:\n%s", err, string(yml2))
		return nil, err
	}

	var res interface{}
	if obj1 == nil {
		res = obj2
	} else if obj2 == nil {
		res = obj1
	} else {
		map1, ok := obj1.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("parse yaml fail, please check file format:\n%s", string(yml1))
		}
		map2, ok := obj2.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("parse yaml fail, please check file format:\n%s", string(yml2))
		}
		res = MergeMap(map1, map2)
	}
	//fmt.Printf("get map:%v", res)
	out, err := yaml.Marshal(res)
	if err != nil {
		return nil, err
	}
	return out, nil
}

//MergeMap maps from first to second
func MergeMap(first map[interface{}]interface{}, second map[interface{}]interface{}) map[interface{}]interface{} {
	if first == nil {
		return second
	}
	if second == nil {
		second = make(map[interface{}]interface{})
	}

	for k, v := range first {
		//fmt.Printf("first type:%v", reflect.TypeOf(first[k]))
		//fmt.Printf("second type:%v", reflect.TypeOf(second[k]))
		if reflect.TypeOf(second[k]) != reflect.TypeOf(first[k]) {
			second[k] = v
			//fmt.Printf("v:%v\n", v)
		} else if reflect.TypeOf(first[k]) == reflect.TypeOf(map[interface{}]interface{}{}) {
			//merge maps
			second[k] = MergeMap(first[k].(map[interface{}]interface{}), second[k].(map[interface{}]interface{}))
		} else {
			//for other types,replace it with value in first map.
			second[k] = v
			//fmt.Printf("cover,v:%v\n", v)
		}
	}

	return second
}
