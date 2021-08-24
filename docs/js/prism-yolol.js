Prism.languages.yolol = {
    'comment': /\/\/.+/,
    'string': /"[^"]*"/,
    'keyword': /((?<=^|\\s|[^a-zA-Z_:])(if|then|else|end|goto)+)|(?<=^|\\s|[^a-zA-Z0-9_:.])(not|abs|sqrt|sin|cos|tan|asin|acos|atan)(?=[^a-zA-Z0-9_:.]|$)/i,
    'operator': /(?<=[^a-zA-Z0-9_:.])(and|or)(?=[^a-zA-Z0-9_:.])/i,
    'function': /[a-z0-9_]+(?=\()/i,
    'variable': /:[a-zA-Z0-9_:.]+|^[a-zA-Z]+[a-zA-Z0-9_.]*/,
    'constant': /[0-9]+(\.[0-9]+)?/
};
