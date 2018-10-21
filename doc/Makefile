everything: go-ecbpass.1
clean:
	rm -f go-ecbpass.1 go-ecbpass.1.xml

go-ecbpass.1.xml: go-ecbpass.1.txt
	asciidoc -b docbook -d manpage go-ecbpass.1.txt
go-ecbpass.1: go-ecbpass.1.xml
	docbook2man go-ecbpass.1.xml