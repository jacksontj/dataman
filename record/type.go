package record

type Record map[string]interface{}

// Get the value addressed by `nameParts`
func (r Record) Get(nameParts []string) (interface{}, bool) {
	// If no key is given, return nothing
	if nameParts == nil || len(nameParts) <= 0 {
		return nil, false
	}
	val, ok := r[nameParts[0]]
	if !ok {
		return nil, ok
	}

	for _, namePart := range nameParts[1:] {
		typedVal, ok := val.(map[string]interface{})
		if !ok {
			return nil, ok
		}
		val, ok = typedVal[namePart]
		if !ok {
			return nil, ok
		}
	}
	return val, true
}

// Set will set a the value at `nameParts` to `newValue` and return a bool
// on whether it was successful
func (r Record) Set(nameParts []string, newValue interface{}) bool {
	// If there is no key, we cannot set the value
	if len(nameParts) <= 0 {
		return false
	}

	var val interface{}
	var ok bool
	if len(nameParts) > 1 {
		val, ok = r[nameParts[0]]
		if !ok {
			val = make(map[string]interface{})
			r[nameParts[0]] = val
		}
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			switch valTyped := val.(type) {
			case map[string]interface{}:
				val, ok = valTyped[namePart]
				if !ok {
					val = make(map[string]interface{})
					valTyped[namePart] = val
				}
			case Record:
				val, ok = valTyped[namePart]
				if !ok {
					val = make(map[string]interface{})
					valTyped[namePart] = val
				}
			default:
				return false
			}
		}

	} else {
		val = r
	}

	switch valTyped := val.(type) {
	case map[string]interface{}:
		valTyped[nameParts[len(nameParts)-1]] = newValue
	case Record:
		valTyped[nameParts[len(nameParts)-1]] = newValue
	default:
		return false
	}

	return true
}

// Remove deletes the value at `nameParts` bool returns whether it is deleted
func (r Record) Remove(nameParts []string) bool {
	// If there is no key, we cannot set the value
	if len(nameParts) <= 0 {
		return false
	}

	var val interface{}
	if len(nameParts) > 1 {
		val = r[nameParts[0]]
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			switch valTyped := val.(type) {
			case map[string]interface{}:
				val = valTyped[namePart]
			case Record:
				val = valTyped[namePart]
			default:
				return false
			}
		}

	} else {
		val = r
	}

	switch valTyped := val.(type) {
	case map[string]interface{}:
		delete(valTyped, nameParts[len(nameParts)-1])
	case Record:
		delete(valTyped, nameParts[len(nameParts)-1])
	default:
		return false
	}
	return true
}

// Pop will pop an item at `nameParts` and return the value and a boolean
// on whether it was successful (similar to map accessing)
func (r Record) Pop(nameParts []string) (interface{}, bool) {
	// If there is no key, we cannot set the value
	if len(nameParts) <= 0 {
		return nil, false
	}

	var val interface{}
	if len(nameParts) > 1 {
		val = r[nameParts[0]]
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			switch valTyped := val.(type) {
			case map[string]interface{}:
				val = valTyped[namePart]
			case Record:
				val = valTyped[namePart]
			default:
				return nil, false
			}
		}

	} else {
		val = r
	}

	switch valTyped := val.(type) {
	case map[string]interface{}:
		tmp, ok := valTyped[nameParts[len(nameParts)-1]]
		if ok {
			delete(valTyped, nameParts[len(nameParts)-1])
		}
		return tmp, ok
	case Record:
		tmp, ok := valTyped[nameParts[len(nameParts)-1]]
		if ok {
			delete(valTyped, nameParts[len(nameParts)-1])
		}
		return tmp, ok
	default:
		return nil, false
	}
}

// Flatten a record into 1-level map[string]interface{}
func (r Record) Flatten() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range r {
		switch typedV := v.(type) {
		// TODO: remove? Or should everything be this? Or do we want a "Subrecord" type
		case map[string]interface{}:
			subMap := Record(typedV).Flatten()
			for subK, subV := range subMap {
				result[k+"."+subK] = subV
			}
		case Record:
			// get the submap as a flattened thing
			subMap := typedV.Flatten()
			for subK, subV := range subMap {
				result[k+"."+subK] = subV
			}
		default:
			result[k] = v
		}
	}
	return result
}

func (r Record) Project(projectionFields [][]string) Record {
	projectedResult := make(Record)
	if projectionFields == nil || len(projectionFields) == 0 {
		return projectedResult
	}
PROJECTLOOP:
	for _, fieldNameParts := range projectionFields {
		switch len(fieldNameParts) {
		case 0:
			continue
		case 1:
			tmpVal, ok := r[fieldNameParts[0]]
			if ok {
				projectedResult[fieldNameParts[0]] = tmpVal
			}
		default:
			dstTmp := projectedResult
			srcTmp := r
			for _, fieldNamePart := range fieldNameParts[:len(fieldNameParts)-1] {
				var ok bool
				srcTmp, ok = srcTmp[fieldNamePart].(map[string]interface{})
				// If the field isn't there in the source, continue on
				if !ok {
					continue PROJECTLOOP
				}
				_, ok = dstTmp[fieldNamePart]
				if !ok {
					dstTmp[fieldNamePart] = make(map[string]interface{})
				}
				dstTmp = dstTmp[fieldNamePart].(map[string]interface{})
			}
			// Now we are on the last hop-- just copy the value over
			tmpVal, ok := srcTmp[fieldNameParts[len(fieldNameParts)-1]]
			if ok {
				dstTmp[fieldNameParts[len(fieldNameParts)-1]] = tmpVal
			}
		}
	}
	return projectedResult
}
