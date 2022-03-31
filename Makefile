init:
	cd tools; go generate -x
	bin/task init

%:
	# Use `bin/task $@` instead.
	bin/task $@
