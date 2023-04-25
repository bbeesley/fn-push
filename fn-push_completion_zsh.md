## fn-push completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(fn-push completion zsh)

To load completions for every new session, execute once:

#### Linux:

	fn-push completion zsh > "${fpath[1]}/_fn-push"

#### macOS:

	fn-push completion zsh > $(brew --prefix)/share/zsh/site-functions/_fn-push

You will need to start a new shell for this setup to take effect.


```
fn-push completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [fn-push completion](fn-push_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 12-Apr-2023