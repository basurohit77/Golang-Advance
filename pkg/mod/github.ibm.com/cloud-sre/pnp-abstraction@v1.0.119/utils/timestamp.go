package utils

import (
	"errors"
	"strconv"
	"time"
)

// VerifyAndConvertTimestamp Verify that the timestamp specified in create_start and create_end are one of these format:
//    yyyy-MM-ddTHH:mm:ssZ
//    yyyy-MM-ddTHH:mm:ss.sssZ
//    yyyy-MM-ddTHH:mm:ss-hhmm
//    yyyy-MM-ddTHH:mm:ss+hhmm
//
// And convert:
//		yyyy-MM-ddTHH:mm:ss-hhmm	to 	yyyy-MM-ddTHH:mm:ssZ UTC timestamp by adding 		hhmm 	to yyyy-MM-ddTHH:mm:ss
//		yyyy-MM-dd HH:mm:ss-hh 		to 	yyyy-MM-ddTHH:mm:ssZ UTC timestamp by adding 		hh 		to yyyy-MM-ddTHH:mm:ss and change space to T
//		yyyy-MM-ddTHH:mm:ss+hhmm	to 	yyyy-MM-ddTHH:mm:ssZ UTC timestamp by substracting 	hhmm 	from yyyy-MM-ddTHH:mm:ss
//		yyyy-MM-dd HH:mm:ss+hh		to 	yyyy-MM-ddTHH:mm:ssZ UTC timestamp by substracting 	hh 		from yyyy-MM-ddTHH:mm:ss and change space to T
func VerifyAndConvertTimestamp(timestamp string) (utcTimestamp string, err error) {
	utcTimestamp = timestamp

	var ts string
	var t time.Time
	if len(timestamp) < 20 {
		err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
		return utcTimestamp, err
	}

	if timestamp[len(timestamp)-1:] == "Z" {
		_, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
			return utcTimestamp, err
		}
	} else if timestamp[len(timestamp)-5:len(timestamp)-4] == "-" {
		ts = timestamp[:len(timestamp)-5] + "Z"
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
			return utcTimestamp, err
		} else {
			h, err1 := strconv.Atoi(timestamp[len(timestamp)-4 : len(timestamp)-2])
			m, err2 := strconv.Atoi(timestamp[len(timestamp)-2:])
			if err1 == nil && err2 == nil {
				after := t.Add(time.Hour*time.Duration(h) + time.Minute*time.Duration(m))
				utcTimestamp = after.Format("2006-01-02T15:04:05Z")
				return utcTimestamp, nil
			} else {
				err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
				return utcTimestamp, err
			}
		}
	} else if timestamp[10:11] == " " && timestamp[len(timestamp)-3:len(timestamp)-2] == "-" {
		ts = timestamp[:10] + "T" + timestamp[11:len(timestamp)-3] + "Z"
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
			return utcTimestamp, err
		} else {
			h, err1 := strconv.Atoi(timestamp[len(timestamp)-2 : len(timestamp)])
			if err1 == nil {
				after := t.Add(time.Hour * time.Duration(h))
				utcTimestamp = after.Format("2006-01-02T15:04:05Z")
				return utcTimestamp, err1
			} else {
				err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
				return utcTimestamp, err
			}
		}
	} else if timestamp[len(timestamp)-5:len(timestamp)-4] == "+" {
		ts = timestamp[:len(timestamp)-5] + "Z"
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
			return utcTimestamp, err
		} else {
			h, err1 := strconv.Atoi(timestamp[len(timestamp)-4 : len(timestamp)-2])
			m, err2 := strconv.Atoi(timestamp[len(timestamp)-2:])
			if err1 == nil && err2 == nil {
				after := t.Add(-time.Hour*time.Duration(h) - time.Minute*time.Duration(m))
				utcTimestamp = after.Format("2006-01-02T15:04:05Z")
				return utcTimestamp, nil
			} else {
				err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
				return utcTimestamp, err
			}
		}
	} else if timestamp[10:11] == " " && timestamp[len(timestamp)-3:len(timestamp)-2] == "+" {
		ts = timestamp[:10] + "T" + timestamp[11:len(timestamp)-3] + "Z"
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
			return utcTimestamp, err
		} else {
			h, err1 := strconv.Atoi(timestamp[len(timestamp)-2 : len(timestamp)])
			if err1 == nil {
				after := t.Add(-time.Hour * time.Duration(h))
				utcTimestamp = after.Format("2006-01-02T15:04:05Z")
				return utcTimestamp, err1
			} else {
				err = errors.New("Error: timestamp is in wrong format, timestamp=[" + timestamp + "]")
				return utcTimestamp, err
			}
		}
	}

	return utcTimestamp, nil
}
