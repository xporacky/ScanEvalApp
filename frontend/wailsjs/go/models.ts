export namespace main {
	
	export class ExamStats {
	    examId: number;
	    studentCount: number;
	    avgScore: number;
	    minScore: number;
	    maxScore: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.examId = source["examId"];
	        this.studentCount = source["studentCount"];
	        this.avgScore = source["avgScore"];
	        this.minScore = source["minScore"];
	        this.maxScore = source["maxScore"];
	    }
	}
	export class ExamSummary {
	    id: number;
	    title: string;
	    schoolYear: string;
	    // Go type: time
	    date: any;
	    questionCount: number;
	    optionCount: number;
	    studentCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.schoolYear = source["schoolYear"];
	        this.date = this.convertValues(source["date"], null);
	        this.questionCount = source["questionCount"];
	        this.optionCount = source["optionCount"];
	        this.studentCount = source["studentCount"];
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
	export class ExamTemplate {
	    title: string;
	    schoolYear: string;
	    dateTime: string;
	    questionCount: number;
	    optionCount: number;
	    showName: boolean;
	    answers: string[];
	    studentCSVContent: string;
	
	    static createFrom(source: any = {}) {
	        return new ExamTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.schoolYear = source["schoolYear"];
	        this.dateTime = source["dateTime"];
	        this.questionCount = source["questionCount"];
	        this.optionCount = source["optionCount"];
	        this.showName = source["showName"];
	        this.answers = source["answers"];
	        this.studentCSVContent = source["studentCSVContent"];
	    }
	}
	export class StudentSummary {
	    id: number;
	    name: string;
	    surname: string;
	    // Go type: time
	    birthDate: any;
	    room: string;
	    registrationNumber: number;
	    examId: number;
	    score: number;
	    pages: string;
	
	    static createFrom(source: any = {}) {
	        return new StudentSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.surname = source["surname"];
	        this.birthDate = this.convertValues(source["birthDate"], null);
	        this.room = source["room"];
	        this.registrationNumber = source["registrationNumber"];
	        this.examId = source["examId"];
	        this.score = source["score"];
	        this.pages = source["pages"];
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

