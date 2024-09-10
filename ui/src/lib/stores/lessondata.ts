// src/stores/auth.ts
import { writable } from 'svelte/store';

export interface LessonInfo {
    discipline: string
    topic: string
    teacher: string
    assignments?: string[]
    day: string
    nextDates?: string[]
}


export const lessons = writable<LessonInfo[]|undefined>(undefined);
