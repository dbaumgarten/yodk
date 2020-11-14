Prism.languages.nolol = {
    'comment': /\/\/.+/,
    'string': /"[^"]*"/,
    'tag': /[a-zA-Z]+[a-zA-Z0-9_]*>/,
    'keyword': /\b(if|then|else|end|goto|define|while|do|wait|include|macro|insert|break|continue|_if|_goto)\b/i,
    'operator': /\b(and|or|not)\b/i,
    'function': /[a-z0-9_]+(?=\()/i,
    'variable': /:?[a-zA-Z]+[a-zA-Z0-9_]*/,
    'constant': /[0-9]+(\.[0-9]+)?/
};
