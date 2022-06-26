Prism.languages.nolol = {
    'comment': /\/\/.+/,
    'string': /"[^"]*"/,
    'tag': /^\s*[a-zA-Z]+[a-zA-Z0-9_]*>/,
    'keyword': /\b(if|then|else|end|goto|define|while|do|include|macro|break|continue|block|line|expr)\b/i,
    'operator': /\b(and|or|not)\b/i,
    'function': /[a-z0-9_]+(?=\()/i,
    'variable': /:[a-zA-Z0-9_:.]+|^[a-zA-Z]+[a-zA-Z0-9_.]*/,
    'constant': /(([0-9]+(\.[0-9]+)?)e[0-9]+)|(0x[0-9a-fA-F]+)|(([0-9]+(\.[0-9]+)?))/
};
