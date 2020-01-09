FILES :=trw.go
TEST_FILES := trw_test.go

TEST := test-result

test: $(TEST)

$(TEST): $(FILES) $(TEST_FILES)
	goimports -w $?
	golint $?
	go test | tee $@

clean:
	rm -f $(TEST)
