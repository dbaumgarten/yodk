Prism.languages.nolol = {
    'comment': /\/\/.+/,
    'string': /"[^"]*"/,
    'tag': /[a-zA-Z]+[a-zA-Z0-9_]*>/,
    'keyword': /\b(if|then|else|end|goto|const|while|do|wait)\b/,
    'operator': /\b(and|or|not)\b/,
    'function': /[a-z0-9_]+(?=\()/i,
    'variable': /:?[a-zA-Z]+[a-zA-Z0-9_]*/,
    'constant': /[0-9]+(\.[0-9]+)?/
};
