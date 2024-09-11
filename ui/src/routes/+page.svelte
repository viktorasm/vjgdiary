<script lang="ts">
    import { goto } from '$app/navigation';
    import { type LessonInfo, lessons } from '$lib/stores/lessondata';
    import {writable} from "svelte/store";
    import moment from 'moment';
    import { onMount } from 'svelte';
    import axios from 'axios';

    export function formatRelativeDate(date: Date): string {
        return moment(date).fromNow();
    }

    const lessonsByDiscipline = writable<Map<LessonInfo[keyof LessonInfo],LessonInfo[]>>()

    function groupBy<T>(array: T[], key: keyof T): Map<T[keyof T], T[]> {
        return array.reduce((result, item) => {
            const groupKey = item[key];
            if (!result.has(groupKey)) {
                result.set(groupKey, []);
            }
            result.get(groupKey)!.push(item);
            return result;
        }, new Map<T[keyof T], T[]>());
    }

    lessons.subscribe(lessonsValue => {
        if (lessonsValue){
            const grouped = groupBy(lessonsValue, "discipline")
            grouped.forEach((value, key) => {
                value.sort((a, b) => {
                    return -a.day.localeCompare(b.day)
                })
            })

            lessonsByDiscipline.set(grouped)
        } else {
            lessonsByDiscipline.set(new Map<LessonInfo[keyof LessonInfo], LessonInfo[]>())
        }
    })

    let loggedIn: any = null

    let loading = false


    onMount(async () =>{
        if (!loggedIn){
            try {
                const response = await axios.get("/api/login");
                loggedIn=response.data
            } catch (e) {

            }
        }

        if (!loggedIn) {
            // Redirect to the login page
            console.log("redirecting to login")
            goto('/login');
            return;
        }


        loading = true
        try {
            const lessonsData = await axios.get("/api/lesson-info")
            lessons.set(lessonsData.data)
        } finally {
            loading = false
        }
    })

    const handleLogout = async (event: Event) => {
        event.preventDefault();
        loggedIn = null;
        lessons.set(undefined);
        await axios.post("/api/logout")
        goto('/login');
    }

    const isDifferenceLessThanADay = (date1: Date|null, date2: Date): boolean =>  {
        // Get the timestamps for both dates
        if (!date1) {
            return false
        }
        const time1 = date1.getTime();
        const time2 = date2.getTime();

        // Calculate the absolute difference
        const differenceInMilliseconds = Math.abs(time1 - time2);

        // 1 day in milliseconds = 24 * 60 * 60 * 1000
        const oneDayInMilliseconds = 24 * 60 * 60 * 1000;

        // Check if the difference is less than 1 day
        return differenceInMilliseconds < oneDayInMilliseconds;
    }
</script>

{#if loggedIn}
<div class="flex flex-col items-center justify-center md:p-6 md:py-8">
    <div class="w-full bg-white shadow dark:border md:mt-0  md:p-6 p-2 dark:bg-gray-800 dark:border-gray-700">

    <h1>{loggedIn.name}</h1>
    <a href="logout" on:click={handleLogout}>Logout</a>
    <div class="max-w-screen-2xl">
        <ul role="list" class="divide-y divide-gray-200 dark:divide-gray-700">
            {#if loading}
                <em>Kraunasi...</em>
            {/if}
        {#each $lessonsByDiscipline as [lessonDiscipline, lessons]}
            {@const nextDate = lessons[0].nextDates?new Date(lessons[0].nextDates[0]):null }
            {@const isNextDay = isDifferenceLessThanADay(nextDate, new Date()) }
            <li class="py-3 sm:py-4">
                <div class="mb-3">
                    <p class="font-medium text-xl text-cyan-900 truncate dark:text-white">{lessonDiscipline} <span class="ml-2 text-gray-500 text-sm">{lessons[0].teacher}</span></p>
                </div>

                {#if nextDate}
                    <p class="mb-3"><span class="text-sm">Kita pamoka:</span> <span class:text-yellow-600={isNextDay} >{nextDate.toLocaleDateString("lt")} {nextDate.toLocaleTimeString("lt")} ({formatRelativeDate(nextDate)})</span></p>
                {/if}

                <p class="mt-2 mb-1 text-sm">Buvusios pamokos:</p>

                <div class="">
                    {#each lessons as lesson, index}
                        {@const day = new Date(lesson.day) }

                        <div class="{index==0?'text-md':'text-sm text-gray-600'} mb-2">
                            <p ><span class="text-xs text-gray-500">{day.toLocaleDateString("lt")} ({formatRelativeDate(day)})</span></p>
                            <div class="flex flex-row">
                                <div class="pr-2 justify-start flex-1 dark:text-red-500">
                                    {lesson.topic}
                                </div>

                                <div class="justify-start flex-1">
                                    {#if lesson.assignments}
                                        <ul class="list-none">
                                            {#each lesson.assignments as assignment}
                                                <li>Užduotis: {assignment}</li>
                                            {/each}
                                        </ul>
                                    {/if}
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>

            </li>
        {/each}
        </ul>
    </div>
</div>
</div>
{/if}