# DailyBurnFinal
more complete than the 3 hour version

Currently I take a ~60 second startup hit to preprocess all of the csv files into memory. After that the api's that I have (and theoretically the rest when written) can simply be lookups and calculations. I prefer to take a one time startup cost so that performance can be faster for all future calls.

### SETUP/Execution

Simply put the unziped csv files into the same directory as the source, build/run
it uses port 12345 on the machine to serve the API's

There are also a few print statments that tell you what step it is currently at while preprocessing. Once it completes it also prints out how long it took to do the preprocessing.

### API calls
`get localhost:12345/session/{session_id}/AllHRM`

returns the min/max/avg bpm for the specific session

`get localhost:12345/session/AllHRM`

returns the min/max/avg bpm for all sessions in json.
I did not implement pagination/result limiting, so it will be a big response.

`get localhost:12345/session/{id}/zones`

returns all important Zone data for the given ID in json.

`get localhost:12345/session/HRZones/newest/{num}`

returns all heartrate information and Zone data for {num} newest sessions, in json.



##### Unrequested API's
`get localhost:12345/people/{id}`

returns json for person {id}

`get localhost:12345/people`

returns all of the people from users.csv in json
I did not implement pagination/result limiting, so it will be a big response.
