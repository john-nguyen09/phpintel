import { Class } from "../symbol/class/class";
import { TypeName } from "../type/name";
import { Serializer, Deserializer } from "./serializer";
import { Symbol } from "../symbol/symbol";
import { Function } from "../symbol/function/function";
import { Variable } from "../symbol/variable/variable";
import { Parameter } from "../symbol/variable/parameter";
import { Constant } from "../symbol/constant/constant";
import { ClassConstant } from "../symbol/constant/classConstant";
import { Method } from "../symbol/function/method";
import { Property } from "../symbol/variable/property";
import { DefineConstant } from "../symbol/constant/defineConstant";

interface SymbolEncoding {
    encode(symbol: Symbol): string;
    decode(buffer: string): Symbol;
}

const classEncoding =  {
    encode: (symbol: Class): string => {
        let serializer = new Serializer();

        serializer.setTypeName(symbol.name);
        serializer.setTypeName(symbol.extend);
        serializer.setLocation(symbol.location);
        serializer.setSymbolModifier(symbol.modifier);

        serializer.setInt32(symbol.implements.length);
        for (let name of symbol.implements) {
            serializer.setTypeName(name);
        }

        serializer.setInt32(symbol.traits.length);
        for (let name of symbol.traits) {
            serializer.setTypeName(name);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: string): Class => {
        let deserializer = new Deserializer(buffer);
        let theClass = new Class();

        theClass.name = deserializer.readTypeName() || new TypeName('');
        theClass.extend = deserializer.readTypeName();
        theClass.location = deserializer.readLocation();
        theClass.modifier = deserializer.readSymbolModifier();

        let noImplements = deserializer.readInt32();
        for (let i = 0; i < noImplements; i++) {
            theClass.implements.push(deserializer.readTypeName() || new TypeName(''));
        }

        let noTraits = deserializer.readInt32();
        for (let i = 0; i < noTraits; i++) {
            theClass.traits.push(deserializer.readTypeName() || new TypeName(''));
        }

        return theClass;
    }
}

const functionEncoding = {
    encode: (symbol: Function): string => {
        let serializer = new Serializer();

        serializer.setTypeName(symbol.name);
        serializer.setLocation(symbol.location);
        serializer.setString(symbol.description);

        serializer.setInt32(symbol.types.length);
        for (let type of symbol.types) {
            serializer.setTypeName(type);
        }

        serializer.setInt32(Object.keys(symbol.scopeVar.variables).length);
        for (let varName in symbol.scopeVar.variables) {
            let variable = symbol.scopeVar.variables[varName];

            serializer.setString(varName);
            serializer.setTypeComposite(variable);
        }

        serializer.setInt32(symbol.parameters.length);
        for (let param of symbol.parameters) {
            serializer.setString(param.name);
            serializer.setTypeComposite(param.type);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: string): Function => {
        let func = new Function();
        let deserializer = new Deserializer(buffer);

        func.name = deserializer.readTypeName() || new TypeName('');
        func.location = deserializer.readLocation();
        func.description = deserializer.readString();

        let noTypes = deserializer.readInt32();
        for (let i = 0; i < noTypes; i++) {
            func.typeAggregate.push(deserializer.readTypeName() || new TypeName(''));
        }

        let noVariables = deserializer.readInt32();
        for (let i = 0; i < noVariables; i++) {
            func.scopeVar.set(new Variable(deserializer.readString(), deserializer.readTypeComposite()));
        }

        let noParams = deserializer.readInt32();
        for (let i = 0; i < noParams; i++) {
            let param = new Parameter();

            param.name = deserializer.readString();
            param.type = deserializer.readTypeComposite();
            func.parameters.push(param);
        }

        return func;
    }
};

const constEncoding = {
    DEFINE_CONSTANT: 1,
    CONSTANT: 2,
    encode: (symbol: Constant): string => {
        let serializer = new Serializer();

        serializer.setInt32(symbol instanceof DefineConstant ?
            constEncoding.DEFINE_CONSTANT : constEncoding.CONSTANT);
        serializer.setTypeName(symbol.name);
        serializer.setLocation(symbol.location);
        serializer.setTypeName(symbol.type);
        serializer.setString(symbol.value);

        return serializer.getBuffer();
    },
    decode: (buffer: string): Constant => {
        let constant: Constant;
        let deserializer = new Deserializer(buffer);
        let constType = deserializer.readInt32();

        if (constType == constEncoding.DEFINE_CONSTANT) {
            constant = new DefineConstant();
        } else {
            constant = new Constant();
        }

        constant.name = deserializer.readTypeName() || new TypeName('');
        constant.location = deserializer.readLocation();
        constant.resolvedType = deserializer.readTypeName();
        constant.resolvedValue = deserializer.readString();

        return constant;
    }
};

const classConstEncoding = {
    encode: (symbol: ClassConstant): string => {
        let serializer = new Serializer();

        serializer.setTypeName(symbol.name);
        serializer.setLocation(symbol.location);
        serializer.setTypeName(symbol.type);
        serializer.setTypeName(symbol.scope);
        serializer.setString(symbol.value);

        return serializer.getBuffer();
    },
    decode: (buffer: string): ClassConstant => {
        let classConst = new ClassConstant();
        let deserializer = new Deserializer(buffer);

        classConst.name = deserializer.readTypeName() || new TypeName('');
        classConst.location = deserializer.readLocation();
        classConst.type = deserializer.readTypeName() || new TypeName('');
        classConst.scope = deserializer.readTypeName();
        classConst.value = deserializer.readString();

        return classConst;
    }
};

const methodEncoding = {
    encode: (symbol: Method): string => {
        let serializer = new Serializer();

        serializer.setTypeName(symbol.name);
        serializer.setLocation(symbol.location);
        serializer.setString(symbol.description);
        serializer.setSymbolModifier(symbol.modifier);
        serializer.setTypeName(symbol.scope);

        serializer.setInt32(symbol.types.length);
        for (let type of symbol.types) {
            serializer.setTypeName(type);
        }

        serializer.setInt32(Object.keys(symbol.variables).length);
        for (let varName in symbol.variables) {
            serializer.setString(varName);
            serializer.setTypeComposite(symbol.variables[varName]);
        }

        serializer.setInt32(symbol.parameters.length);
        for (let param of symbol.parameters) {
            serializer.setString(param.name);
            serializer.setTypeComposite(param.type);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: string): Method => {
        let method = new Method();
        let deserializer = new Deserializer(buffer);

        method.name = deserializer.readTypeName() || new TypeName('');
        method.location = deserializer.readLocation();
        method.description = deserializer.readString();
        method.modifier = deserializer.readSymbolModifier();
        method.scope = deserializer.readTypeName();

        let noTypes = deserializer.readInt32();
        for (let i = 0; i < noTypes; i++) {
            method.pushType(deserializer.readTypeName());
        }

        let noVariables = deserializer.readInt32();
        for (let i = 0; i < noVariables; i++) {
            method.setVariable(new Variable(deserializer.readString(), deserializer.readTypeComposite()));
        }

        const noParams = deserializer.readInt32();
        for (let i = 0; i < noParams; i++) {
            const param = new Parameter();
            param.name = deserializer.readString();
            param.type = deserializer.readTypeComposite();
            method.pushParam(param);
        }

        return method;
    }
};

const propertyEncoding = {
    encode: (symbol: Property): string => {
        let serializer = new Serializer();

        serializer.setString(symbol.name);
        serializer.setLocation(symbol.location);
        serializer.setString(symbol.description);
        serializer.setSymbolModifier(symbol.modifier);
        serializer.setTypeName(symbol.scope);
        serializer.setTypeComposite(symbol.type);

        return serializer.getBuffer();
    },
    decode: (buffer: string): Property => {
        let property = new Property();
        let deserializer = new Deserializer(buffer);

        property.name = deserializer.readString();
        property.location = deserializer.readLocation();
        property.description = deserializer.readString();
        property.modifier = deserializer.readSymbolModifier();
        property.scope = deserializer.readTypeName();
        property.type = deserializer.readTypeComposite();

        return property;
    }
};

enum Type {
    Class = 1,
    Function = 2,
    Constant = 3,
    ClassConstant = 4,
    Method = 5,
    Property = 6,
};

const symbolEncodingMap: {[key: number]: SymbolEncoding} = {
    1: classEncoding,
    2: functionEncoding,
    3: constEncoding,
    4: classConstEncoding,
    5: methodEncoding,
    6: propertyEncoding,
};

export = {
    type: 'symbol-encoding',
    encode: (symbol: Symbol): string => {
        let serializer = new Serializer();
        let symbolType: Type | null = null;

        if (symbol instanceof Class) {
            symbolType = Type.Class;
        } else if (symbol instanceof Function) {
            symbolType = Type.Function;
        } else if (symbol instanceof Constant) {
            symbolType = Type.Constant;
        } else if (symbol instanceof ClassConstant) {
            symbolType = Type.ClassConstant;
        } else if (symbol instanceof Method) {
            symbolType = Type.Method;
        } else if (symbol instanceof Property) {
            symbolType = Type.Property;
        }

        if (symbolType !== null) {
            serializer.setInt32(symbolType);

            let encoding = symbolEncodingMap[symbolType];
            serializer.setBuffer(encoding.encode(symbol));
        }

        return serializer.getBuffer();
    },
    decode: (buffer: string): Symbol => {
        if (typeof buffer !== 'string' || buffer.length == 0) {
            throw new Error(`Invalid buffer`);
        }

        let deserializer = new Deserializer(buffer);
        let type = deserializer.readInt32();
        buffer = deserializer.readBuffer();

        if (type in symbolEncodingMap) {
            return symbolEncodingMap[type].decode(buffer);
        }

        throw new Error(`Invalid buffer`);
    },
    buffer: false
} as Level.Encoding;
