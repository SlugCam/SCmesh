package distribute

/*
func TestRegistration(t *testing.T) {
	dir, err := ioutil.TempDir("", "SCmesh_test")
	if err != nil {
		t.Fatal("error making tmp directory to run test in.")
	}

	//var message *json.RawMessage
	//err = json.Unmarshal([]byte("{\"test\":40}"), message)
	message := json.RawMessage(`{"test":45}`)
	ackCh := make(chan ACK)
	d, err := Distribute(dir, ackCh)
	if err != nil {
		t.Fatal("Error in Distribute:", err)
	}

	_, err = d.Register(RegistrationRequest{
		DataType:    "message",
		Destination: uint32(0),
		Timestamp:   time.Now(),
		JSON:        &message,
	})
	if err != nil {
		t.Fatal("Error in Register:", err)
	}
	time.Sleep(2 * time.Second)
}
*/
