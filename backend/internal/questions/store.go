package questions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("question not found")

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Upsert(ctx context.Context, input ImportInput) (Question, bool, error) {
	questionText := normalizeText(input.Question)
	_, optionsHash := normalizeOptions(input.Options)
	answerText := normalizeAnswer(input.Answer, input.Options)
	questionHash := hashText(questionText)
	answerHash := hashText(answerText)
	compositeHash := hashText(strings.Join([]string{questionHash, optionsHash, answerHash}, ":"))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Question{}, false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO questions
		(question_text, normalized_text, question_hash, options_hash, answer_hash, composite_hash, question_type, platform, subject, source, collected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, questionText, questionText, questionHash, optionsHash, answerHash, compositeHash,
		defaultType(input.Type), strings.TrimSpace(input.Platform), strings.TrimSpace(input.Subject), strings.TrimSpace(input.Source), input.CollectedAt)
	if err != nil {
		return Question{}, false, fmt.Errorf("insert question: %w", err)
	}
	created := false
	if affected, _ := result.RowsAffected(); affected > 0 {
		created = true
	}
	var questionID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM questions WHERE composite_hash = ?`, compositeHash).Scan(&questionID); err != nil {
		return Question{}, false, err
	}
	if created {
		for position, option := range input.Options {
			if _, err := tx.ExecContext(ctx, `INSERT INTO question_options (question_id, option_key, option_text, position) VALUES (?, ?, ?, ?)`, questionID, strings.TrimSpace(option.Key), normalizeText(option.Text), position); err != nil {
				return Question{}, false, err
			}
		}
		answers := strings.Split(answerText, "###")
		raw := input.AnswerRaw
		if raw == "" {
			raw = input.Answer
		}
		for position, answer := range answers {
			if _, err := tx.ExecContext(ctx, `INSERT INTO question_answers (question_id, answer_text, answer_raw, position) VALUES (?, ?, ?, ?)`, questionID, answer, raw, position); err != nil {
				return Question{}, false, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return Question{}, false, err
	}
	question, err := s.GetByID(ctx, questionID)
	return question, created, err
}

func (s *Store) Search(ctx context.Context, query string) (Question, error) {
	question, _, err := s.SearchWithScore(ctx, query)
	return question, err
}

func (s *Store) SearchWithScore(ctx context.Context, query string, optionValues ...[]string) (Question, float64, error) {
	normalized := normalizeText(query)
	if normalized == "" {
		return Question{}, 0, ErrNotFound
	}
	var id uint64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM questions WHERE status = 1 AND question_hash = ? ORDER BY id ASC LIMIT 1`, hashText(normalized)).Scan(&id)
	if err == nil {
		question, getErr := s.GetByID(ctx, id)
		question.Similarity = 1
		return question, 1, getErr
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = s.db.QueryRowContext(ctx, `SELECT id FROM questions WHERE status = 1 AND INSTR(normalized_text, ?) > 0 ORDER BY CHAR_LENGTH(normalized_text), id ASC LIMIT 1`, normalized).Scan(&id)
		if err == nil {
			question, getErr := s.GetByID(ctx, id)
			score := similarityScore(normalized, normalizeText(question.Question))
			question.Similarity = score
			return question, score, getErr
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		matchText := normalized
		if len(optionValues) > 0 && len(optionValues[0]) > 0 {
			matchText += " " + normalizeText(strings.Join(optionValues[0], " "))
		}
		question, score, candidateErr := s.searchSimilar(ctx, normalized, matchText)
		if candidateErr == nil && score >= 0.35 {
			question.Similarity = score
			return question, score, nil
		}
		return Question{}, 0, ErrNotFound
	}
	if err != nil {
		return Question{}, 0, err
	}
	return Question{}, 0, ErrNotFound
}

func (s *Store) searchSimilar(ctx context.Context, normalized, matchText string) (Question, float64, error) {
	terms := similarityTerms(matchText)
	if len(terms) == 0 {
		return Question{}, 0, ErrNotFound
	}
	conditions := make([]string, 0, len(terms))
	args := make([]any, 0, len(terms))
	for _, term := range terms {
		conditions = append(conditions, "INSTR(normalized_text, ?) > 0")
		args = append(args, term)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT q.id, q.normalized_text, COALESCE((SELECT GROUP_CONCAT(qo.option_text SEPARATOR ' ') FROM question_options qo WHERE qo.question_id = q.id), '') FROM questions q WHERE q.status = 1 AND (`+strings.Join(conditions, " OR ")+") ORDER BY q.id DESC LIMIT 100", args...)
	if err != nil {
		return Question{}, 0, err
	}
	defer rows.Close()
	var bestID uint64
	bestScore := 0.0
	for rows.Next() {
		var id uint64
		var candidate, candidateOptions string
		if err := rows.Scan(&id, &candidate, &candidateOptions); err != nil {
			return Question{}, 0, err
		}
		if candidateOptions != "" {
			candidate += " " + normalizeText(candidateOptions)
		}
		score := similarityScore(matchText, candidate)
		if score > bestScore {
			bestID, bestScore = id, score
		}
	}
	if err := rows.Err(); err != nil {
		return Question{}, 0, err
	}
	if bestID == 0 {
		return Question{}, 0, ErrNotFound
	}
	question, err := s.GetByID(ctx, bestID)
	return question, bestScore, err
}

func similarityTerms(value string) []string {
	fields := strings.Fields(value)
	terms := make([]string, 0, 12)
	if len(fields) > 1 {
		for _, field := range fields {
			if len([]rune(field)) >= 2 {
				terms = append(terms, field)
			}
			if len(terms) >= 12 {
				break
			}
		}
		return terms
	}
	runes := []rune(value)
	for index := 0; index+1 < len(runes) && len(terms) < 12; index++ {
		terms = append(terms, string(runes[index:index+2]))
	}
	return terms
}

func similarityScore(left, right string) float64 {
	left = normalizeText(left)
	right = normalizeText(right)
	if left == "" || right == "" {
		return 0
	}
	if left == right {
		return 1
	}
	leftSet := ngramSet(left)
	rightSet := ngramSet(right)
	if len(leftSet) == 0 || len(rightSet) == 0 {
		return 0
	}
	intersection := 0
	for value := range leftSet {
		if _, ok := rightSet[value]; ok {
			intersection++
		}
	}
	union := len(leftSet) + len(rightSet) - intersection
	if union == 0 {
		return 0
	}
	score := float64(intersection) / float64(union)
	if strings.Contains(left, right) || strings.Contains(right, left) {
		lengthRatio := float64(minInt(len([]rune(left)), len([]rune(right)))) / float64(maxInt(len([]rune(left)), len([]rune(right))))
		if lengthRatio > score {
			score = lengthRatio
		}
	}
	return score
}

func ngramSet(value string) map[string]struct{} {
	runes := []rune(value)
	set := make(map[string]struct{}, len(runes))
	if len(runes) == 1 {
		set[value] = struct{}{}
		return set
	}
	for index := 0; index+1 < len(runes); index++ {
		set[string(runes[index:index+2])] = struct{}{}
	}
	return set
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func (s *Store) ListAdmin(ctx context.Context, search, questionType, subject string, status, page, pageSize int) (QuestionPage, error) {
	page, pageSize = normalizePage(page, pageSize)
	where := "WHERE 1 = 1"
	args := make([]any, 0, 8)
	if strings.TrimSpace(search) != "" {
		where += " AND q.normalized_text LIKE ?"
		args = append(args, "%"+normalizeText(search)+"%")
	}
	if strings.TrimSpace(questionType) != "" {
		where += " AND q.question_type = ?"
		args = append(args, strings.TrimSpace(questionType))
	}
	if strings.TrimSpace(subject) != "" {
		where += " AND q.subject = ?"
		args = append(args, strings.TrimSpace(subject))
	}
	if status == 0 || status == 1 {
		where += " AND q.status = ?"
		args = append(args, status)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM questions q "+where, args...).Scan(&total); err != nil {
		return QuestionPage{}, err
	}
	args = append(args, (page-1)*pageSize, pageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT q.id, q.question_text, q.question_type, q.platform, q.subject,
		q.source, q.status, q.collected_at, q.created_at,
		(SELECT COUNT(*) FROM question_options qo WHERE qo.question_id = q.id),
		(SELECT COUNT(*) FROM question_answers qa WHERE qa.question_id = q.id)
		FROM questions q `+where+` ORDER BY q.id DESC LIMIT ?, ?`, args...)
	if err != nil {
		return QuestionPage{}, err
	}
	defer rows.Close()
	items := make([]QuestionSummary, 0)
	for rows.Next() {
		var item QuestionSummary
		if err := rows.Scan(&item.ID, &item.Question, &item.Type, &item.Platform, &item.Subject,
			&item.Source, &item.Status, &item.CollectedAt, &item.CreatedAt, &item.OptionCount, &item.AnswerCount); err != nil {
			return QuestionPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return QuestionPage{}, err
	}
	return QuestionPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Store) UpdateStatus(ctx context.Context, id uint64, status int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE questions SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetByID(ctx context.Context, id uint64) (Question, error) {
	var question Question
	err := s.db.QueryRowContext(ctx, `SELECT id, question_text, question_type, platform, subject, source, status, collected_at FROM questions WHERE id = ?`, id).
		Scan(&question.ID, &question.Question, &question.Type, &question.Platform, &question.Subject, &question.Source, &question.Status, &question.CollectedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Question{}, ErrNotFound
	}
	if err != nil {
		return Question{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT option_key, option_text, position FROM question_options WHERE question_id = ? ORDER BY position ASC`, id)
	if err != nil {
		return Question{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.Key, &option.Text, &option.Position); err != nil {
			return Question{}, err
		}
		question.Options = append(question.Options, option)
	}
	answers, err := s.db.QueryContext(ctx, `SELECT answer_text, answer_raw, position FROM question_answers WHERE question_id = ? ORDER BY position ASC`, id)
	if err != nil {
		return Question{}, err
	}
	defer answers.Close()
	for answers.Next() {
		var answer Answer
		if err := answers.Scan(&answer.Text, &answer.Raw, &answer.Position); err != nil {
			return Question{}, err
		}
		question.Answers = append(question.Answers, answer)
	}
	return question, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func defaultType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "other"
	}
	return value
}
