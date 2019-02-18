import { TypeName } from "../type/name";
import { Location } from "../symbol/meta/location";
import { Range } from "../symbol/meta/range";
import { SymbolModifier } from "../symbol/meta/modifier";
import { TypeComposite } from "../type/composite";
import { NamespaceName } from "../symbol/name/namespaceName";

export class Serializer {
    private buffer: string;

    constructor() {
        this.buffer = '';
    }

    public setBuffer(buffer: string) {
        this.buffer += buffer;
    }

    public setString(str: string) {
        this.setInt32(str.length);
        this.setBuffer(str);
    }

    public setInt32(value: number) {
        this.buffer += String.fromCharCode((value >> 24 & 255)) +
            String.fromCharCode((value >> 16 & 255)) +
            String.fromCharCode((value >> 8 & 255)) +
            String.fromCharCode((value & 255));
    }

    public setBool(value: boolean) {
        this.buffer += value ? String.fromCharCode(1) : String.fromCharCode(0);
    }

    public setTypeComposite(types: TypeComposite) {
        this.setInt32(types.types.length);
        for (let type of types.types) {
            this.setTypeName(type);
        }
    }

    public setTypeName(name: TypeName | null) {
        if (name == null) {
            this.setBool(false);
        } else {
            this.setBool(true);
            this.setString(name.name);
        }
    }

    public setLocation(location: Location) {
        if (location.uri === undefined || location.range === undefined) {
            this.setBool(false);
        } else {
            this.setBool(true);
            this.setString(location.uri);
            this.setRange(location.range);
        }
    }

    public setRange(range: Range) {
        this.setInt32(range.start);
        this.setInt32(range.end);
    }

    public setSymbolModifier(modifier: SymbolModifier) {
        this.setInt32(modifier.getModifier());
        this.setInt32(modifier.getVisibility());
    }

    public setNamespaceName(namespace: NamespaceName) {
        this.setString(namespace.parts.join('\\'));
    }

    public getBuffer(): string {
        return this.buffer;
    }
}

export class Deserializer {
    private buffer: string;
    private offset: number;

    constructor(buffer: string) {
        this.buffer = buffer;
        this.offset = 0;
    }

    public readBuffer(): string {
        return this.buffer.slice(this.offset);
    }

    public readString(): string {
        const length = this.readInt32();
        const str = this.buffer.slice(this.offset, this.offset + length);
        this.offset += length;

        return str;
    }

    public readInt32(): number {
        const value = this.buffer.charCodeAt(this.offset + 0) << 24 |
            this.buffer.charCodeAt(this.offset + 1) << 16 |
            this.buffer.charCodeAt(this.offset + 2) << 8 |
            this.buffer.charCodeAt(this.offset + 3);
        this.offset += 4;

        return value;
    }

    public readBool(): boolean {
        let value = this.buffer.charCodeAt(this.offset) == 1 ? true : false;
        this.offset += 1;

        return value;
    }

    public readTypeComposite(): TypeComposite {
        let numTypes = this.readInt32();
        let types = new TypeComposite();
        for (let i = 0; i < numTypes; i++) {
            types.push(this.readTypeName() || new TypeName(''));
        }

        return types;
    }

    public readTypeName(): TypeName | null {
        let hasTypeName = this.readBool();

        if (!hasTypeName) {
            return null;
        }

        return new TypeName(this.readString());
    }

    public readLocation(): Location {
        let hasLocation = this.readBool();

        if (!hasLocation) {
            return {};
        }

        return { uri: this.readString(), range: this.readRange() };
    }

    public readRange(): Range {
        return {
            start: this.readInt32(),
            end: this.readInt32()
        }
    }

    public readSymbolModifier(): SymbolModifier {
        return new SymbolModifier(this.readInt32(), this.readInt32());
    }

    public readNamespaceName(): NamespaceName {
        let namespace = new NamespaceName();

        namespace.parts = this.readString().split('\\');

        return namespace;
    }
}
