// Package validation provides framework-independent request validation and
// translated validation errors based on go-playground/validator.
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Rule describes a custom validator rule and its translated error message.
type Rule struct {
	Tag      string
	Message  string
	Validate validator.Func
}

// Option customizes a Validator.
type Option func(*config)

type config struct {
	fieldTagName string
	rules        []Rule
}

// WithFieldTagName selects the struct tag used as the translated field name.
// It defaults to json. This does not change the validation tag name.
func WithFieldTagName(tagName string) Option {
	return func(cfg *config) {
		cfg.fieldTagName = tagName
	}
}

// WithRules appends project-specific validation rules.
func WithRules(rules ...Rule) Option {
	return func(cfg *config) {
		cfg.rules = append(cfg.rules, rules...)
	}
}

// Validator validates structs and translates validation errors into Chinese.
// It does not depend on an HTTP framework.
type Validator struct {
	engine     *validator.Validate
	translator ut.Translator
}

// Error is the first translated field validation error.
type Error struct {
	Field   string
	Tag     string
	Message string
	Cause   error
}

// Error returns the translated validation message.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap exposes the original validation error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// New creates an independent validator that reads rules from validate tags.
func New(options ...Option) (*Validator, error) {
	return Configure(validator.New(), options...)
}

// Configure registers common rules and Chinese translations on an existing
// validator engine. It leaves the engine's validation tag name unchanged, so
// callers can use validate, binding, or another tag selected by the framework.
func Configure(engine *validator.Validate, options ...Option) (*Validator, error) {
	if engine == nil {
		return nil, errors.New("validator engine must not be nil")
	}

	cfg := config{
		fieldTagName: "json",
		rules:        defaultRules(),
	}
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	registerFieldName(engine, cfg.fieldTagName)
	chineseLocale := zh.New()
	translator, found := ut.New(chineseLocale, chineseLocale).GetTranslator("zh")
	if !found {
		return nil, errors.New("Chinese validator translator is unavailable")
	}
	if err := zhTranslations.RegisterDefaultTranslations(engine, translator); err != nil {
		return nil, fmt.Errorf("register default Chinese translations: %w", err)
	}
	for _, rule := range cfg.rules {
		if err := registerRule(engine, translator, rule); err != nil {
			return nil, err
		}
	}

	return &Validator{
		engine:     engine,
		translator: translator,
	}, nil
}

// Engine returns the configured go-playground validator engine.
func (v *Validator) Engine() *validator.Validate {
	return v.engine
}

// Validate validates a struct and returns its first translated error.
func (v *Validator) Validate(value any) error {
	if v == nil || v.engine == nil {
		return errors.New("validator is not initialized")
	}
	return v.Translate(v.engine.Struct(value))
}

// Translate converts validator.ValidationErrors, including wrapped errors,
// into a deterministic first translated error. Other errors pass through.
func (v *Validator) Translate(err error) error {
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return err
	}
	fieldError := validationErrors[0]
	return &Error{
		Field:   fieldError.Field(),
		Tag:     fieldError.Tag(),
		Message: fieldError.Translate(v.translator),
		Cause:   err,
	}
}

// FirstMessage returns the translated first validation message.
func (v *Validator) FirstMessage(err error) string {
	translatedErr := v.Translate(err)
	if translatedErr == nil {
		return ""
	}
	return translatedErr.Error()
}

var (
	defaultOnce      sync.Once
	defaultValidator *Validator
	defaultErr       error
)

// Default returns the lazily initialized package validator.
func Default() (*Validator, error) {
	defaultOnce.Do(func() {
		defaultValidator, defaultErr = New()
	})
	return defaultValidator, defaultErr
}

// Validate validates a struct with the package default validator.
func Validate(value any) error {
	defaultInstance, err := Default()
	if err != nil {
		return err
	}
	return defaultInstance.Validate(value)
}

// FirstMessage translates an error with the package default validator.
func FirstMessage(err error) string {
	defaultInstance, initErr := Default()
	if initErr != nil {
		return initErr.Error()
	}
	return defaultInstance.FirstMessage(err)
}

func registerFieldName(engine *validator.Validate, tagName string) {
	engine.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get(tagName), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return field.Name
		}
		return name
	})
}

func registerRule(engine *validator.Validate, translator ut.Translator, rule Rule) error {
	if strings.TrimSpace(rule.Tag) == "" {
		return errors.New("validation rule tag must not be empty")
	}
	if rule.Validate == nil {
		return fmt.Errorf("validation rule %q has no validator function", rule.Tag)
	}
	if err := engine.RegisterValidation(rule.Tag, rule.Validate); err != nil {
		return fmt.Errorf("register validation rule %q: %w", rule.Tag, err)
	}
	if err := engine.RegisterTranslation(
		rule.Tag,
		translator,
		func(translator ut.Translator) error {
			return translator.Add(rule.Tag, rule.Message, false)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			message, err := translator.T(fieldError.Tag(), fieldError.Field())
			if err != nil {
				return fieldError.Error()
			}
			return message
		},
	); err != nil {
		return fmt.Errorf("register translation for rule %q: %w", rule.Tag, err)
	}
	return nil
}
