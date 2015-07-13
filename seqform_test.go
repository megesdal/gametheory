package gametheory

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSequenceFormString(t *testing.T) {
	gameJSON := []byte(`{
    "player": "A",
    "iset": "A1",
    "moves": [{
      "name": "L",
      "next": {
        "player": "B",
        "iset": "B1",
        "moves":[{
          "name": "a",
          "outcome": [{
            "player": "A",
            "payoff": 11
          },{
            "player": "B",
            "payoff": 3
          }]
        },{
          "name": "b",
          "outcome": [{
            "player": "A",
            "payoff": 3
          },{
            "player": "B",
            "payoff": 0
          }]
        }]
      }
    },{
      "name":"R",
      "next": {
			  "player":"A",
				"iset":"A2",
				"moves":[{
		      "name":"S",
		      "next": {
					  "player":"!",
		        "chance": true,
		        "moves":[{
						  "name":"?",
		          "prob":0.5,
		          "next":{
		            "player":"B",
		            "iset":"B1",
		            "moves":[{
		              "name":"a",
		              "outcome":[{
		                "player":"A",
		                "payoff":0
		              },{
		                "player":"B",
		                "payoff":0
		              }]
		            },{
		              "name":"b",
		              "outcome":[{
		                "player":"A",
		                "payoff":0
		              },{
		                "player":"B",
		                "payoff":10
		              }]
		            }]
		          }
		        },{
		          "name":"?",
							"prob":0.5,
		          "next":{
		            "player":"B",
		            "iset":"B2",
		            "moves":[{
		              "name":"c",
		              "outcome":[{
		                "player":"A",
		                "payoff":0
		              },{
		                "player":"B",
		                "payoff":4
		              }]
		            },{
		              "name":"d",
		              "outcome":[{
		                "player":"A",
		                "payoff":24
		              },{
		                "player":"B",
		                "payoff":0
		              }]
		            }]
		          }
		        }]
		      }
		    },{
					"name":"T",
					"next":{
						"player":"B",
		        "iset":"B2",
		        "moves":[{
		          "name":"c",
		          "outcome":[{
		            "player":"A",
		            "payoff":6
		          },{
		            "player":"B",
		            "payoff":0
		          }]
		        },{
		          "name":"d",
		          "outcome":[{
		            "player":"A",
		            "payoff":0
		          },{
		            "player":"B",
		            "payoff":1
		          }]
		        }]
		      }
				}]
      }
    }]
  }`)

	var rootFactory NodeFactory
	err := json.Unmarshal(gameJSON, &rootFactory)
	if err != nil {
		t.Error(err)
	}

	sf := NewSequenceForm(&rootFactory)
	fmt.Println(sf.String())

	sf.CreateLCP()
}
