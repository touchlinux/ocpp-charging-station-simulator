package usecases

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

type Validator struct {
	request []byte
	matches map[string]interface{}
}

type ValidatorList []Validator
type ValidatorMap map[string]ValidatorList

func (r Validator) GetRequest() []byte {
	return r.request
}

func match(jsonMap map[string]interface{}, matchKey string, matchVal interface{}) bool {
	fmt.Printf("jsonMap has %d items\n", len(jsonMap))
	for jsonKey, jsonVal := range jsonMap {
		fmt.Printf("Matching {%s:%s} with {%s:%s}\n", jsonKey, jsonVal, matchKey, matchVal)
		fmt.Println("type:", reflect.TypeOf(jsonVal))
		var jsonValMap map[string]interface{}
		if reflect.TypeOf(jsonVal) == reflect.TypeOf(jsonValMap) {
			// Parsed value has object type
			jsonValMap = jsonVal.(map[string]interface{})
			fmt.Println("Value is an object :", jsonVal, ", len:", len(jsonValMap))
			if match(jsonValMap, matchKey, matchVal) == true {
				return true
			}
		} else if matchKey == jsonKey {
			// Parsed value has string type
			if reflect.ValueOf(matchVal).String() == reflect.ValueOf(jsonVal).String() {
				fmt.Printf("Matched with {%s:%s}.\n", jsonKey, jsonVal)
				return true
			} else {
				fmt.Printf("Mismatched. {%s:%s} is not {%s:%s}.\n", jsonKey, jsonVal, matchKey, matchVal)
			}
		}
	}
	return false
}

func (r Validator) Validate(response []byte) bool {
	call := &wrappers.CALL{}
	if err := call.UnmarshalJSON(r.request); err != nil {
		fmt.Println(err)
		return false
	}
	callresult := &wrappers.CALLRESULT{}
	if err := callresult.UnmarshalJSON(response); err != nil {
		fmt.Println(err)
		return false
	}

	// Invalid if MessageId mismatch
	if call.MessageId != callresult.MessageId {
		fmt.Println("MessageId is mismatched!")
		return false
	}

	if r.matches == nil {
		// No matches, no validation
		return true
	}
	// Iterate matches to check all of them can be found from the response
	for matchKey, matchVal := range r.matches {
		//fmt.Printf("matchKey:%s, matchVal:%s\n", matchKey, matchVal)
		var jsonObj interface{}
		json.Unmarshal(callresult.GetPayloadAsJSON(), &jsonObj)
		jsonMap := jsonObj.(map[string]interface{})
		// All matches have to be satisfied
		if match(jsonMap, matchKey, matchVal) == false {
			return false
		}
	}
	return true
}
