Prism.languages.yolol = {
    'comment': /\/\/.+/,
    'string': /"[^"]*"/,
    'keyword': /\b(if|then|else|end|goto)\b/i,
    'operator': /\b(and|or|not)\b/i,
    'function': /[a-z0-9_]+(?=\()/i,
    'variable': /:?[a-zA-Z]+[a-zA-Z0-9_]*/,
    'constant': /[0-9]+(\.[0-9]+)?/
};
