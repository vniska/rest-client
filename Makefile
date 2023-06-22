test_php:
	@echo "\n\033[92m\033[1mPHP\033[0m\033[0m"
	./vendor/bin/phpunit --colors --verbose --bootstrap ./tests/php/phpunit.php ./tests/php/

test_python:
	@echo "\n\033[92m\033[1mPYTHON\033[0m\033[0m"
	python3 -m unittest discover tests/python/ -p '*_test.py' -v

test_golang:
	@echo "\n\033[92m\033[1mGOLANG\033[0m\033[0m"
	go test ./golang/...

test:
	make test_php
	make test_python
	make test_golang
