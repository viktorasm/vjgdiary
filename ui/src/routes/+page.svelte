<script lang="ts">
    import { goto } from '$app/navigation';
    import { type LessonInfo, lessons } from '$lib/stores/lessondata';
    import {writable} from "svelte/store";
    import moment from 'moment';
    import { onMount } from 'svelte';
    import axios from 'axios';
    import Title from "$lib/components/title.svelte";
    import _ from 'lodash';




    export function formatRelativeDate(date: Date): string {
        return moment(date).fromNow();
    }

    type LessonsByCategory = {
        category: string
        lessons: LessonInfo[]
    }
    type Discipline = {
        name: string
        nextDate?:Date
        teachers:string[]
        categories: LessonsByCategory[]
    }

    let compactView = true;



    type LessonsByDisciplineAndCategory = Discipline[]
    const lessonsByDisciplineAndCategory = writable<LessonsByDisciplineAndCategory>()

    function mapToGroupedLessons(lessons: LessonInfo[]): LessonsByDisciplineAndCategory {
        const now = new Date()

        const disciplines = _.map(_.groupBy(lessons, "discipline"),  (lessons: LessonInfo[], discipline: string) : Discipline=> {
            const sortedLessons = lessons.sort((a, b) => {
                return b.day.getTime() - a.day.getTime()
            })
            const [before, after] = _.partition(sortedLessons, (item: LessonInfo) => new Date(item.day) < now);

            const categories = [] as LessonsByCategory[]

            if (compactView) {
                categories.push({
                    category: "Paskutinė pamoka",
                    lessons: before.splice(0,1),
                })
            } else {
                if (before && before.length>0) {
                    before[0].isNextForThisDiscipline = true;
                    categories.push({
                        category: "Praėjusios pamokos",
                        lessons: compactView?before.splice(0,1):before,
                    })
                }
                if (after && after.length>0) {
                    categories.push({
                        category: "Suplanuotos pamokos",
                        lessons: after,
                    })
                }
            }


            return {
                name: discipline,
                teachers: _.uniq(_.map(lessons, (l):string => {
                    return l.teacher
                })),
                nextDate: lessons[0].nextDates?.[0],
                categories: categories,
            }
        })

        return disciplines;
    }

    lessons.subscribe(lessonsValue => {
        if (lessonsValue){
            const result = mapToGroupedLessons(lessonsValue)

            lessonsByDisciplineAndCategory.set(result)
        } else {
            lessonsByDisciplineAndCategory.set([])
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
            const lessonsData = await axios.get("/api/lesson-info");
            const items = lessonsData.data;
            items.forEach((i:any) => {
                if (i.day) {
                    i.day = new Date(i.day)
                }
                if (i.nextDates) {
                    i.nextDates = i.nextDates.map((i:any)=>{
                        return new Date(i)
                    })
                }
            })
            lessons.set(items)
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

    // returns true if date2 is same or next day for date1
    const isNextDay = (date1?: Date, date2?: Date): boolean =>  {
        if (!date1 || !date2) {
            return false
        }
        const startOfDay = new Date(date1);
        startOfDay.setHours(0, 0, 0, 0);

        // return true for next two days since start of day for date1
        return Math.abs(startOfDay.getTime() - date2.getTime()) < 24 * 60 * 60 * 1000 * 2;
    }
    const formatDate = (date: Date): string => {
        const result = date.toLocaleString("lt")
        if (result.endsWith(":00")){
            return result.slice(0, -3)
        }
        return result
    }
</script>

<Title title=""/>


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
        {#each $lessonsByDisciplineAndCategory as discipline}
            {@const nextDay = isNextDay(discipline.nextDate, new Date()) }
            <li class="py-3 sm:py-4">
                <div class="mb-2">
                    <p class="font-medium text-xl text-cyan-900 truncate dark:text-white">{discipline.name} <span class="ml-2 text-gray-500 text-sm">{discipline.teachers}</span></p>
                </div>

                {#if discipline.nextDate}
                    <p class="mb-1"><span class="text-sm">Kita pamoka:</span> <span class:text-yellow-600={nextDay} >{formatDate(discipline.nextDate)} ({formatRelativeDate(discipline.nextDate)})</span></p>
                {/if}

                {#each discipline.categories as category}
                    <div class="pb-3 pt-2">
                    <p class="mt-2 mb-1 text-md text-gray-600 font-medium">{category.category}</p>

                    <div class="">
                        {#each category.lessons as lesson}
                            {@const day = new Date(lesson.day) }

                            <div class="{lesson.isNextForThisDiscipline?'text-md':'text-sm text-gray-600'} mb-3">
                                <p ><span class="text-xs text-gray-500">{formatDate(day)} ({formatRelativeDate(day)})</span></p>

                                {#if lesson.lessonNotes}
                                    <div class="pt-2 pb-3"><span class="font-bold">{lesson.lessonNotes.category}</span> <span class="bg-amber-100 px-3 py-1 rounded-full">{lesson.lessonNotes.note}</span></div>
                                {/if}
                                {#if lesson.mark}
                                    <div class="font-bold">Pažymys: <span class="bg-amber-500 text-white text-sm font-bold px-3 py-1 rounded-full">{lesson.mark}</span></div>
                                {/if}

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
                    </div>
                {/each}

            </li>
        {/each}
        </ul>
    </div>
</div>
</div>
{/if}