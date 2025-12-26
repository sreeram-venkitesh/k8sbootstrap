package certificates

func SetupCerts() error {
	err := createKubernetesCA()
	if err != nil {
		return err
	}

	return nil
}
