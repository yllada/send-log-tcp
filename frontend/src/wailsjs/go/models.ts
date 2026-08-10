export namespace main {
	
	export class BatchImportResult {
	    messages: string[];
	    totalLines: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new BatchImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messages = source["messages"];
	        this.totalLines = source["totalLines"];
	        this.errors = source["errors"];
	    }
	}
	export class ConnectionProfile {
	    id: string;
	    name: string;
	    description?: string;
	    address: string;
	    port: string;
	    protocol: string;
	    framingMethod: string;
	    useTls: boolean;
	    tlsVerify: boolean;
	    caCertPath?: string;
	    clientCertPath?: string;
	    clientKeyPath?: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.address = source["address"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.framingMethod = source["framingMethod"];
	        this.useTls = source["useTls"];
	        this.tlsVerify = source["tlsVerify"];
	        this.caCertPath = source["caCertPath"];
	        this.clientCertPath = source["clientCertPath"];
	        this.clientKeyPath = source["clientKeyPath"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ConnectionService {
	
	
	    static createFrom(source: any = {}) {
	        return new ConnectionService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class ContinuousSendConfig {
	    Address: string;
	    Port: string;
	    Protocol: string;
	    Message: string;
	    FramingMethod: string;
	    Facility: number;
	    Severity: number;
	    Hostname: string;
	    Appname: string;
	    MessageFormat: string;
	    UseRFC5424: boolean;
	    UseTLS: boolean;
	    TLSVerify: boolean;
	    CACertPath: string;
	    ClientCertPath: string;
	    ClientKeyPath: string;
	    Duration: number;
	    MessagesPerSec: number;
	    MaxMessages: number;
	    RandomizeData: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ContinuousSendConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.Port = source["Port"];
	        this.Protocol = source["Protocol"];
	        this.Message = source["Message"];
	        this.FramingMethod = source["FramingMethod"];
	        this.Facility = source["Facility"];
	        this.Severity = source["Severity"];
	        this.Hostname = source["Hostname"];
	        this.Appname = source["Appname"];
	        this.MessageFormat = source["MessageFormat"];
	        this.UseRFC5424 = source["UseRFC5424"];
	        this.UseTLS = source["UseTLS"];
	        this.TLSVerify = source["TLSVerify"];
	        this.CACertPath = source["CACertPath"];
	        this.ClientCertPath = source["ClientCertPath"];
	        this.ClientKeyPath = source["ClientKeyPath"];
	        this.Duration = source["Duration"];
	        this.MessagesPerSec = source["MessagesPerSec"];
	        this.MaxMessages = source["MaxMessages"];
	        this.RandomizeData = source["RandomizeData"];
	    }
	}
	export class ContinuousStats {
	    totalSent: number;
	    totalErrors: number;
	    currentRate: number;
	    elapsedSeconds: number;
	    isRunning: boolean;
	    startTime: number;
	    targetRate: number;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new ContinuousStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSent = source["totalSent"];
	        this.totalErrors = source["totalErrors"];
	        this.currentRate = source["currentRate"];
	        this.elapsedSeconds = source["elapsedSeconds"];
	        this.isRunning = source["isRunning"];
	        this.startTime = source["startTime"];
	        this.targetRate = source["targetRate"];
	        this.duration = source["duration"];
	    }
	}
	export class LogTemplate {
	    id: string;
	    name: string;
	    description?: string;
	    message: string;
	    facility: number;
	    severity: number;
	    appname: string;
	    messageFormat?: string;
	    useRfc5424: boolean;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new LogTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.message = source["message"];
	        this.facility = source["facility"];
	        this.severity = source["severity"];
	        this.appname = source["appname"];
	        this.messageFormat = source["messageFormat"];
	        this.useRfc5424 = source["useRfc5424"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class StressTestService {
	
	
	    static createFrom(source: any = {}) {
	        return new StressTestService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SyslogConfig {
	    Address: string;
	    Port: string;
	    Protocol: string;
	    Messages: string[];
	    FramingMethod: string;
	    Facility: number;
	    Severity: number;
	    Hostname: string;
	    Appname: string;
	    MessageFormat: string;
	    UseRFC5424: boolean;
	    UseTLS: boolean;
	    TLSVerify: boolean;
	    CACertPath: string;
	    ClientCertPath: string;
	    ClientKeyPath: string;
	
	    static createFrom(source: any = {}) {
	        return new SyslogConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.Port = source["Port"];
	        this.Protocol = source["Protocol"];
	        this.Messages = source["Messages"];
	        this.FramingMethod = source["FramingMethod"];
	        this.Facility = source["Facility"];
	        this.Severity = source["Severity"];
	        this.Hostname = source["Hostname"];
	        this.Appname = source["Appname"];
	        this.MessageFormat = source["MessageFormat"];
	        this.UseRFC5424 = source["UseRFC5424"];
	        this.UseTLS = source["UseTLS"];
	        this.TLSVerify = source["TLSVerify"];
	        this.CACertPath = source["CACertPath"];
	        this.ClientCertPath = source["ClientCertPath"];
	        this.ClientKeyPath = source["ClientKeyPath"];
	    }
	}
	export class SyslogResponse {
	    sentMessages: string[];
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new SyslogResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sentMessages = source["sentMessages"];
	        this.errors = source["errors"];
	    }
	}
	export class SyslogService {
	
	
	    static createFrom(source: any = {}) {
	        return new SyslogService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

