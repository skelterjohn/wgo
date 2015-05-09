;; Experimental wgo support for EMACS.
;; Allows go-mode and other go tools to be able to handle the
;; GOPATH changes that wgo does when they shell out to go for stuff.
;; To use, add (load-file "/path/to/this/file.el") to your emacs config.

(defun setup-go-mode-env ()
  (make-local-variable 'process-environment)
  (let ((val (shell-command-to-string "wgo env GOPATH")))
    (if (not (string= val "no workspace"))
	(setenv "GOPATH" val)
      )
    )
  )

(add-hook 'go-mode 'setup-go-mode-env)

(defun her-apply-function (orig-fun name)
  (let ((res (funcall orig-fun name)))
    (if (or (string= name "*gocode*") (string= name "*compilation*"))
	(setup-go-mode-env)
      )
    res))

(advice-add 'generate-new-buffer :around #'her-apply-function)
