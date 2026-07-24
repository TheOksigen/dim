package seed

import (
	"fmt"
	"strings"
	"time"
)

const base36 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func FINForIndex(index int) string {
	value := index
	characters := make([]byte, 6)
	for position := len(characters) - 1; position >= 0; position-- {
		characters[position] = base36[value%len(base36)]
		value /= len(base36)
	}
	return "T" + string(characters)
}

func AnswerSequence(seed int) string {
	answers := make([]byte, 30)
	for index := range answers {
		answers[index] = "ABCD"[(seed+index)%4]
	}
	return string(answers)
}

func RowForIndex(index int) []any {
	firstName := fmt.Sprintf("SınaqAd%04d", index%10_000)
	lastName := fmt.Sprintf("DemoSoyad%04d", (index/7)%10_000)
	fatherName := fmt.Sprintf("TestAta%04d", (index/13)%10_000)
	score1 := int16(45 + index%56)
	score2 := int16(42 + (index*3)%59)
	score3 := int16(40 + (index*5)%61)
	total := score1 + score2 + score3
	examDate := time.Date(2026, time.Month(index%12+1), index%27+1, 0, 0, 0, 0, time.UTC)

	return []any{
		FINForIndex(index), firstName, lastName, fatherName, examDate,
		"Azərbaycan dili", score1, AnswerSequence(index),
		"Riyaziyyat", score2, AnswerSequence(index + 1),
		"İngilis dili", score3, AnswerSequence(index + 2),
		total, total >= 150,
	}
}

func ColumnNames() []string {
	return strings.Fields("fin_code first_name last_name father_name exam_date " +
		"subject_1_name subject_1_score subject_1_answers " +
		"subject_2_name subject_2_score subject_2_answers " +
		"subject_3_name subject_3_score subject_3_answers " +
		"total_score passing_status")
}
