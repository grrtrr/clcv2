#
# Targets. Technically, it should be called 'CLC Console', but that is
# hard to pronounce. Hence 'clonsole' :-)
#
WINDOWS_TARGET := clonsole.exe
LINUX_TARGET   := clconsole

.PHONY: all clean $(LINUX_TARGET)

$(LINUX_TARGET):
	go build -o $@ .
	strip -s $@

$(WINDOWS_TARGET):
	GOOS=windows GOARCH=amd64 go build -o $@ .

clean:
	$(RM) $(LINUX_TARGET) $(WINDOWS_TARGET) *~
