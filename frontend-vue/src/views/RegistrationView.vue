<script setup lang="ts">
import { ref } from 'vue';
import { registration } from '@/api.ts';

const error = ref('')
const successMsg = ref('')
const username = ref('')
const password = ref('')
const role = ref<'engineer' | 'admin'>('engineer')
const adminToken = ref('')
async function handleRegistration()  {
    error.value = ''
    successMsg.value = ''
    console.log(role.value)
    console.log(adminToken.value)
    console.log(role.value === 'admin' ? adminToken.value : undefined)
    
    try {
        await registration(
            username.value,
            password.value,
            role.value,
            role.value === 'admin' ? adminToken.value : undefined
        )
        successMsg.value = "Your account is successfully created!"
    } catch (e) {
        error.value = (e as Error).message ?? 'Registration failed'
    }
}

</script>

<template>
  <main>
    <div class="registration-screen">
        <div class ="registration-card">
            <div class="registration-brand">
                <span class="registration-wordmark">HANDOFF</span>
                <span class="registration-mark">//</span>
            </div>
            <p class="registration-tag">Incident context across shift changes</p>
            <form class="registration-form" @submit.prevent="handleRegistration">
                <p v-if="error" class="error" role="alert">{{ error }}</p>
                <p v-if="successMsg" class="success-message" role="alert">{{ successMsg }}</p>
                <div class="field">
                    <label class="field-label">Username</label>
                    <input class="input" type="text" v-model="username" placeholder="tom" autocomplete="username" required>
                </div>
                <div class="field">
                    <label class="field-label">Password</label>
                    <input class="input" v-model="password" type="password" placeholder="●●●●●●●●" autocomplete="current-password" required>
                    <p class="field-hint">Min 8 characters. Letters, numbers, symbols allowed.</p>
                </div>
                <div class="field">
                    <label class="field-label">Role</label>
                    <div class="radio-group">
                        <label class="radio-option">
                            <input type="radio" v-model="role" value="engineer">
                            <span>Engineer</span>
                        </label>
                        <label class="radio-option">
                            <input type="radio" v-model="role" value="admin">
                            <span>Admin</span>
                        </label>
                    </div>
                </div>
                 <div v-if="role === 'admin'" class="field">
                    <label class="field-label">Admin Token</label>
                    <input class="input" type="password" v-model="adminToken" placeholder="●●●●●●●●" autocomplete="off" required>
                    <p class="field-hint">Admin Token is "adminTokenIncidentHandoff"</p>

                </div>
                <button class="btn btn-primary btn-block" type="submit">Register</button>
            </form>
            <p class="registration-foot mono dim">ON-CALL ACCESS ONLY</p>

            <RouterLink :to="{name:'log-in'}" class="back mono">← Login</RouterLink>
        </div>
    </div>
  </main>
</template>

<style scoped>
/* Delete this CSS */
.field-hint {
    font-size: 11px;
    color: var(--color-text-dim);
    margin-top: 4px;
}
.back {
    padding-top: 30px;
    display: flex;
    justify-content: center;
    font-size: 15px;
}

.registration-screen {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 90vh;
    padding: 24px;
}

.registration-card {
    background-color: var(--color-panel);
    border: 1px solid var(--color-border);
    border-top: 3px solid var(--color-accent);
    border-radius: 8px;
    padding: 40px;
    max-width: 380px;
    width: 100%;    
}

.registration-brand{
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    gap: 5px;
}

.registration-wordmark {
    color: var(--color-text-bright);
    font-family: var(--font-mono);
    font-size: 26px;
    font-weight: 700;
    letter-spacing: 4px;
}

.registration-tag {
    color: var(--color-text-dim);
    font-size: 13px;
    margin: 12px 0 28px;
    text-align: center;
}

.registration-mark {
    color: var(--color-accent);
    font-size: 26px;
}

.registration-form {
    margin-bottom: 20px;
}

.radio-group {
    display: flex;
    gap: 24px;
    margin-top: 6px;
}
.radio-option {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: var(--color-text);
}
.radio-option input[type="radio"] {
    accent-color: var(--color-accent);
    width: 15px;
    height: 15px;
}

.registration-foot{
    font-size: 10px;
    letter-spacing: 2px;
    text-align: center;
}
</style>