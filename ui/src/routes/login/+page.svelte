<script lang="ts">
    import {goto} from "$app/navigation";
    import axios from 'axios';


    let username = '';
    let password = '';

    let loading = false;
    let errorMessage = "";


    const handleSubmit = async (event: Event) => {
        event.preventDefault();
        loading = true;

        try {
            errorMessage = ""
            const req = {
                "username": username.toString(),
                "password": password.toString(),
            }
            const response = await axios.post("/api/login", req);

            goto("/")
        } catch (error) {
            errorMessage = "Nepavyko prisijungti."
            console.log(error)
        }

        loading = false;


    };
</script>

<div class="flex flex-col items-center justify-center px-6 py-8 mx-auto  md:h-screen">
    <div class="w-full bg-white rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">


<div class="p-6 space-y-4 md:space-y-6 sm:p-8 " class:display-none={loading}>
    <h1>VJG dienynas</h1>
    <form class="space-y-4 md:space-y-6" on:submit|preventDefault={handleSubmit}>
        <div>
            <label for="userid" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white"  >Mokinio ID</label>
            <input type="number" name="userid" id="userid" bind:value={username} class="bg-gray-50 border border-gray-300 text-gray-900 rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="12345" required="">
        </div>
        <div>
            <label for="password" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white" >Slaptažodis</label>
            <input type="password" name="password" id="password" placeholder="••••••••" bind:value={password} class="bg-gray-50 border border-gray-300 text-gray-900 rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" required="">
        </div>
        {#if errorMessage}
            <div>
                <p class="text-amber-700">{errorMessage}</p>
            </div>
        {/if}
        <div class="flex items-center justify-between">
            <div class="flex items-start">
                <!--                            <div class="flex items-center h-5">-->
                <!--                                <input id="remember" aria-describedby="remember" type="checkbox" class="w-4 h-4 border border-gray-300 rounded bg-gray-50 focus:ring-3 focus:ring-primary-300 dark:bg-gray-700 dark:border-gray-600 dark:focus:ring-primary-600 dark:ring-offset-gray-800" required="">-->
                <!--                            </div>-->
                <!--                            <div class="ml-3 text-sm">-->
                <!--                                <label for="remember" class="text-gray-500 dark:text-gray-300">Remember me</label>-->
                <!--                            </div>-->
            </div>
        </div>
        <button type="submit" class="w-full text-white bg-primary-600 hover:bg-primary-700 focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800">Prisijungti</button>
    </form>
</div>


</div>
</div>