package options

func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.InsecureServing.Validate()...)
	errs = append(errs, o.Log.Validate()...)
	errs = append(errs, o.Sqlite.Validate()...)

	return errs
}
