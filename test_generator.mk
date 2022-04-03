bmg_test_postgres_backup:
	./.build/bmg backup \
		--definition=generate/test_data/examples/postgres.yaml \
		--template postgres

bmg_test_postgres_backup_k8s:
	./.build/bmg backup \
		--definition=generate/test_data/examples/postgres.yaml \
		--template postgres \
		--kubernetes \
		--gpg-key-path generate/test_data/examples/gpg.key

bmg_test_postgres_backup_k8s_sealed_secret:
	./.build/bmg backup \
		--definition=generate/test_data/examples/postgres.yaml \
		--template postgres \
		--kubernetes \
		--gpg-key-path generate/test_data/examples/valid-sealed-secret.yaml
