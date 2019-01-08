export namespace WordSeparator {
    export function getTokens(name: string): string[] {
        let lastSlash = name.lastIndexOf('\\');

        if (lastSlash >= 0) {
            name = name.substr(lastSlash + 1);
        }
        
        if (name.indexOf('_') >= 0) {
            return underscore(name);
        } else {
            return casing(name);
        }
    }

    function underscore(name: string): string[] {
        let tokens: string[] = [];
        let gotUnderscore = false;

        tokens.push(name);
        for (let i = 0; i < name.length; i++) {
            if (gotUnderscore) {
                if (name[i] === '_') {
                    continue;
                }

                tokens.push(name.substr(i));
                gotUnderscore = false;
            } else {
                if (name[i] === '_') {
                    gotUnderscore = true;
                    continue;
                }
            }
        }

        return tokens;
    }
    
    function casing(name: string): string[] {
        let tokens: string[] = [];
        let isPrevUpper = false;
        let start = -1;

        tokens.push(name);
        for (let i = 0; i < name.length; i++) {
            let isCurrUpper = !isLowerCase(name[i]);

            if (isCurrUpper !== isPrevUpper) {
                if (start == -1) {
                    if (isCurrUpper) {
                        if (i !== 0) {
                            start = i;
                        }
                    } else {
                        if (i !== 1) {
                            tokens.push(name.substr(i - 1));
                        }
                    }
                } else {
                    tokens.push(name.substr(start));

                    if (start != (i - 1) && !isCurrUpper) {
                        tokens.push(name.substr(i - 1));
                    }

                    start = -1;
                }
                
            }

            isPrevUpper = isCurrUpper;
        }
        
        return tokens;
    }

    function isLowerCase(str: string): boolean {
        return str.toLowerCase() == str;
    }
}