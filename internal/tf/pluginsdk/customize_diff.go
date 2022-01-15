package pluginsdk

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type (
	CustomizeDiffFunc        = func(context.Context, *ResourceDiff, interface{}) error
	ValueChangeConditionFunc = func(ctx context.Context, old, new, meta interface{}) bool
)

// CustomDiffWithAll returns a CustomizeDiffFunc that runs all of the given
// CustomizeDiffFuncs and returns all of the errors produced.
//
// If one function produces an error, functions after it are still run.
// If this is not desirable, use function Sequence instead.
//
// If multiple functions returns errors, the result is a multierror.
func CustomDiffWithAll(funcs ...CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		var err error
		for _, f := range funcs {
			thisErr := f(ctx, d, meta)
			if thisErr != nil {
				err = multierror.Append(err, thisErr)
			}
		}
		return err
	}
}

// CustomDiffInSequence returns a CustomizeDiffFunc that runs all of the given
// CustomizeDiffFuncs in sequence, stopping at the first one that returns
// an error and returning that error.
//
// If all functions succeed, the combined function also succeeds.
func CustomDiffInSequence(funcs ...CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		for _, f := range funcs {
			err := f(ctx, d, meta)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
