export namespace main {
	
	export class NoticeLog {
	    text: string;
	    metaData: string;
	    title: string;
	    settingId: string;
	    canCopy: boolean;
	    timestamp: string;
	    isSystem: boolean;
	    isError: boolean;
	    actionSuccess: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NoticeLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.metaData = source["metaData"];
	        this.title = source["title"];
	        this.settingId = source["settingId"];
	        this.canCopy = source["canCopy"];
	        this.timestamp = source["timestamp"];
	        this.isSystem = source["isSystem"];
	        this.isError = source["isError"];
	        this.actionSuccess = source["actionSuccess"];
	    }
	}
	export class SimpleBlock {
	    type: string;
	    start: string;
	    end: string;
	
	    static createFrom(source: any = {}) {
	        return new SimpleBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}
	export class Setting {
	    id: string;
	    title: string;
	    details: string;
	    isCore: boolean;
	    target: string;
	    type: string;
	    url: string;
	    regexp: string;
	    exclude: string;
	    messageKey: string;
	    extraFields: Record<string, string>;
	    simpleMode: boolean;
	    simplePattern: string;
	    simpleBlocks: SimpleBlock[];
	
	    static createFrom(source: any = {}) {
	        return new Setting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.details = source["details"];
	        this.isCore = source["isCore"];
	        this.target = source["target"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.regexp = source["regexp"];
	        this.exclude = source["exclude"];
	        this.messageKey = source["messageKey"];
	        this.extraFields = source["extraFields"];
	        this.simpleMode = source["simpleMode"];
	        this.simplePattern = source["simplePattern"];
	        this.simpleBlocks = this.convertValues(source["simpleBlocks"], SimpleBlock);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveData {
	    path: string;
	    settings: Setting[];
	
	    static createFrom(source: any = {}) {
	        return new SaveData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.settings = this.convertValues(source["settings"], Setting);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

