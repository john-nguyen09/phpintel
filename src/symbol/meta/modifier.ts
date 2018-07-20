export class SymbolModifier {
    static readonly NONE = 0;
    static readonly PUBLIC = 1 << 0;
    static readonly PROTECTED = 1 << 1;
    static readonly PRIVATE = 1 << 2;
    static readonly FINAL = 1 << 3;
    static readonly ABSTRACT = 1 << 4;
    static readonly STATIC = 1 << 5;

    private modifier: number;

    constructor(modifier?: number) {
        if (!modifier) {
            modifier = SymbolModifier.NONE;
        }
        
        this.modifier = modifier;
    }

    has(modifier: number) {
        return (this.modifier & modifier) > 0;
    }

    include(modifier: number) {
        this.modifier |= modifier;
    }
}