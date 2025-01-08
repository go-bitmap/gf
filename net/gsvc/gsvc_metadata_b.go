package gsvc

// Del delete a key-value pair from metadata.
func (m Metadata) Del(key string) {
	delete(m, key)
}
