# Targets
TARGET         := clconsole
WINDOWS_TARGET := $(TARGET).exe
MAC_TARGET     := $(TARGET).mac


.PHONY: all linux windows mac clean get-deps

linux:   $(TARGET)
windows: $(WINDOWS_TARGET)
mac:     $(MAC_TARGET)

# Download dependencies
get-deps:
	go get -d

$(TARGET): get-deps
	go build -o $@ .
	strip -s $@

$(WINDOWS_TARGET): get-deps
	GOOS=windows GOARCH=amd64 go build -o $@ .

$(MAC_TARGET): get-deps
	GOOS=darwin GOARCH=amd64 go build  -o $@ .

clean:
	$(RM) $(TARGET) $(WINDOWS_TARGET) $(MAC_TARGET) *~
