package bobpad

type ApiVariantResponse[TOk any] struct {
	Ok  *TOk    `ic:"Ok,variant"`
	Err *string `ic:"Err,variant"`
}
