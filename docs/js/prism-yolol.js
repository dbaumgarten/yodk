Prism.languages.yolol = {
    'comment': /\/\/.+/,
    'string': /"[^"]*"/,
    'keyword': /(?<=\b|[^a-zA-Z])(if|then|else\b|end\b|goto|abs|sqrt|sin|cos|tan|asin|acos|atan)+/i,
    'operator': /(?<=\b|[^a-zA-Z])(and|or|not)+/i,
    'function': /[a-z0-9_]+(?=\()/i,
    'variable': /:?[a-zA-Z]+[a-zA-Z0-9_]*/,
    'constant': /[0-9]+(\.[0-9]+)?/
};
