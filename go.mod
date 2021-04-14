module github.com/dbaumgarten/yodk

go 1.14

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/fatih/color v1.9.0 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/go-serve/bindatafs v0.0.0-20180828091044-2f268e76aac4 // indirect
	github.com/google/go-dap v0.2.0
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.1
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/yaml.v2 v2.2.7
)

replace github.com/google/go-dap v0.2.0 => github.com/dbaumgarten/go-dap v0.2.2
