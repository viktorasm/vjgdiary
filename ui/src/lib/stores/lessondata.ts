// src/stores/auth.ts
import { writable } from 'svelte/store';



export interface LessonNotes {
    category: string
    note: string
}

export interface LessonInfo {
    discipline: string
    topic: string
    teacher: string
    assignments?: string[]
    day: Date
    nextDates?: Date[]

    isNextForThisDiscipline: boolean

    mark?: string
    lessonNotes?:LessonNotes
}


export const lessons = writable<LessonInfo[]|undefined>(undefined);
