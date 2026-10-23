# 新闻 AI Agent Makefile(参照 stock_ai 风格精简)
BINS=newsserver

.PHONY: build run tidy sql clean

build:
	@for b in $(BINS); do go build -o _output/$$b ./cmd/$$b; done
	@echo "build -> _output/"

run:
	@go run ./cmd/newsserver --config=configs/newsserver.yml

tidy:
	@go mod tidy

sql:
	@mysql -h127.0.0.1 -uroot -p < configs/schema/news_ai_agent.sql

clean:
	@rm -rf _output log
