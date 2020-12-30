package data

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type User struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	Email        string `bson:"email" json:"email"`
	Password     string `bson:"password" json:"password"`
	Name         string `bson:"name" json:"name"`
	RefreshToken string `bson:"refresh_token" json:"-"`
}

type Idea struct {
	ID           string    `bson:"id,omitempty" json:"id,omitempty"`
	Content      string    `bson:"content" json:"content"`
	Impact       int       `bson:"impact" json:"impact"`
	Ease         int       `bson:"ease" json:"ease"`
	Confidence   int       `bson:"confidence" json:"confidence"`
	AverageScore int       `bson:"average_score,omitempty" json:"average_score,omitempty"`
	CreatedAt    time.Time `bson:"created_date" json:"-"`
	TimeStamp    int64     `bson:"created_at" json:"created_at"`
}

func (i *Idea) Validate() error {
	if len(i.Content) == 0 {
		return fmt.Errorf("content is required")
	}
	if len(i.Content) > 255 {
		return fmt.Errorf("content maximum limit reached")
	}
	if err := i.ValidateScore(i.Confidence); err != nil {
		return fmt.Errorf("confidence %v", err)
	}
	if err := i.ValidateScore(i.Ease); err != nil {
		return fmt.Errorf("ease %v", err)
	}
	if err := i.ValidateScore(i.Impact); err != nil {
		return fmt.Errorf("impact %v", err)
	}

	return nil
}
func (i *Idea) ValidateScore(score int) error {
	if !(score >= 1 && score <= 10) {
		return fmt.Errorf("score must be between 1-10")
	}
	return nil
}
func (u *User) Validate() error {
	if len(u.Name) == 0 {
		return fmt.Errorf("name is required")
	}
	if !u.isEmailValid() {
		return fmt.Errorf("invalid email")
	}
	if err := u.verifyPassword(); err != nil {
		return fmt.Errorf("password:%v", err)
	}
	return nil
}
func (u *User) verifyPassword() error {
	var uppercasePresent bool
	var lowercasePresent bool
	var numberPresent bool
	const minPassLength = 8
	const maxPassLength = 64
	var passLen int
	var errorString string

	for _, ch := range u.Password {
		switch {
		case unicode.IsNumber(ch):
			numberPresent = true
			passLen++
		case unicode.IsUpper(ch):
			uppercasePresent = true
			passLen++
		case unicode.IsLower(ch):
			lowercasePresent = true
			passLen++
		case ch == ' ':
			passLen++
		}
	}
	appendError := func(err string) {
		if len(strings.TrimSpace(errorString)) != 0 {
			errorString += ", " + err
		} else {
			errorString = err
		}
	}
	if !lowercasePresent {
		appendError("lowercase letter missing")
	}
	if !uppercasePresent {
		appendError("uppercase letter missing")
	}
	if !numberPresent {
		appendError("at least one numeric character required")
	}
	if !(minPassLength <= passLen && passLen <= maxPassLength) {
		appendError(fmt.Sprintf("password length must be between %d to %d characters long", minPassLength, maxPassLength))
	}

	if len(errorString) != 0 {
		return fmt.Errorf(errorString)
	}
	return nil
}

// isEmailValid checks if the email provided passes the required structure and length.
func (u *User) isEmailValid() bool {
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(u.Email) < 3 && len(u.Email) > 254 {
		return false
	}
	return emailRegex.MatchString(u.Email)
}
