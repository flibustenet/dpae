// Copyright (c) 2023 William Dode. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dpae

import (
	"strings"
	"time"
)

type Date time.Time

func (j *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*j = Date(t)
	return nil
}
func (j Date) String() string {
	return j.Format("2006-01-02")
}
func (j Date) Format(s string) string {
	return time.Time(j).Format(s)
}

type Time time.Time

func (j *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("15:04", s)
	if err != nil {
		return err
	}
	*j = Time(t)
	return nil
}
func (j Time) String() string {
	return j.Format("15:04:05")
}
func (j Time) Format(s string) string {
	return time.Time(j).Format(s)
}
