package data

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

var ErrInvalidDateFormat = errors.New("invalid date format")

type DateOnly time.Time

func (d *DateOnly) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidDateFormat
	}

	parsedTime, err := time.Parse(time.DateOnly, unquotedJSONValue)
	if err != nil {
		return ErrInvalidDateFormat
	}

	*d = DateOnly(parsedTime)

	return nil
}

func (d *DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *DateOnly) String() string {
	return time.Time(*d).Format(time.DateOnly)
}
