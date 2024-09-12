// src/stores/auth.ts
import { writable } from 'svelte/store';


export interface LessonInfo {
    discipline: string
    topic: string
    teacher: string
    assignments?: string[]
    day: Date
    nextDates?: Date[]

    isNextForThisDiscipline: boolean
}


export const lessons = writable<LessonInfo[]|undefined>(undefined);
