// Code generated automaticaly by reset utility; DO NOT EDIT.

package metrics

func (v *Metric) Reset() {

	v.ID = ""

	v.MType = ""

	if v.Delta != nil {
		*v.Delta = 0
	}

	if v.Value != nil {
		*v.Value = 0.0
	}

	v.ValueStr = ""

}
