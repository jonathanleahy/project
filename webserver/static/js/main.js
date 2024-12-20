// Initialize Monaco Editor
require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.43.0/min/vs' }});

let appConfigEditor;
let payloadEditor;

require(['vs/editor/editor.main'], function() {
    // Initialize application config editor
    appConfigEditor = monaco.editor.create(document.getElementById('app-config-editor'), {
        value: JSON.stringify({
            "name": "example-app",
            "path": "/usr/local/bin/processor",
            "args": ["--mode", "fast"],
            "env": {
                "MAX_MEMORY": "1G"
            },
            "pass_payload": true
        }, null, 2),
        language: 'json',
        theme: 'vs',
        minimap: { enabled: false }
    });

    // Initialize payload editor
    payloadEditor = monaco.editor.create(document.getElementById('payload-editor'), {
        value: JSON.stringify({
            "input": "data.csv",
            "output_format": "json",
            "compression": true
        }, null, 2),
        language: 'json',
        theme: 'vs',
        minimap: { enabled: false }
    });
});

// API client
class JobSchedulerAPI {
    constructor(baseURL = '') {
        this.baseURL = baseURL;
    }

    async submitJob(jobData) {
        const response = await fetch(`${this.baseURL}/api/v1/jobs`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(jobData)
        });

        if (!response.ok) {
            throw new Error(`Failed to submit job: ${response.statusText}`);
        }

        return response.json();
    }

    async getChannelStats() {
        const response = await fetch(`${this.baseURL}/api/v1/stats/channels`);
        if (!response.ok) {
            throw new Error(`Failed to get channel stats: ${response.statusText}`);
        }
        return response.json();
    }

    async getActiveJobs() {
        const response = await fetch(`${this.baseURL}/api/v1/jobs?status=running`);
        if (!response.ok) {
            throw new Error(`Failed to get active jobs: ${response.statusText}`);
        }
        return response.json();
    }
}

// UI Controller
class DashboardUI {
    constructor(api) {
        this.api = api;
        this.setupEventListeners();
        this.startPolling();
    }

    setupEventListeners() {
        // Modal controls
        document.getElementById('submit-job').addEventListener('click', () => this.showModal());
        document.getElementById('close-modal').addEventListener('click', () => this.hideModal());
        document.getElementById('submit-modal').addEventListener('click', () => this.handleJobSubmission());

        // Close modal on background click
        document.getElementById('job-modal').addEventListener('click', (e) => {
            if (e.target.id === 'job-modal') {
                this.hideModal();
            }
        });
    }

    showModal() {
        document.getElementById('job-modal').classList.remove('hidden');
    }

    hideModal() {
        document.getElementById('job-modal').classList.add('hidden');
    }

    async handleJobSubmission() {
        try {
            const jobData = {
                job_id: document.getElementById('job-id').value,
                channel: document.getElementById('channel').value,
                workers: parseInt(document.getElementById('workers').value),
                timeout_seconds: parseInt(document.getElementById('timeout').value),
                application: JSON.parse(appConfigEditor.getValue()),
                payload: JSON.parse(payloadEditor.getValue())
            };

            const response = await this.api.submitJob(jobData);
            this.hideModal();
            this.showNotification('Job submitted successfully', 'success');
            this.updateDashboard();
        } catch (error) {
            this.showNotification(`Error submitting job: ${error.message}`, 'error');
        }
    }

    showNotification(message, type) {
        const indicator = document.getElementById('status-indicator');
        indicator.textContent = message;
        indicator.className = `px-3 py-1 rounded-full text-sm ${
            type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
        }`;
        setTimeout(() => {
            indicator.textContent = '';
            indicator.className = 'px-3 py-1 rounded-full text-sm';
        }, 3000);
    }

    async updateDashboard() {
        try {
            const [channelStats, activeJobs] = await Promise.all([
                this.api.getChannelStats(),
                this.api.getActiveJobs()
            ]);

            this.renderChannelStats(channelStats);
            this.renderActiveJobs(activeJobs);
        } catch (error) {
            console.error('Error updating dashboard:', error);
        }
    }

    renderChannelStats(stats) {
        const container = document.getElementById('channel-stats');
        container.innerHTML = Object.entries(stats)
            .map(([channel, stat]) => `
                <div class="border rounded p-4">
                    <h3 class="font-semibold">${channel}</h3>
                    <div class="grid grid-cols-2 gap-2 mt-2 text-sm">
                        <div>Workers: ${stat.workers}</div>
                        <div>Active Jobs: ${stat.active_jobs.length}</div>
                        <div>Total Jobs: ${stat.total_jobs}</div>
                        <div>Failed Jobs: ${stat.failed_jobs}</div>
                    </div>
                </div>
            `)
            .join('');
    }

    renderActiveJobs(jobs) {
        const container = document.getElementById('active-jobs');
        container.innerHTML = jobs.jobs
            .map(job => `
                <div class="border rounded p-4">
                    <div class="flex justify-between items-center">
                        <h3 class="font-semibold">${job.job_id}</h3>
                        <span class="px-2 py-1 rounded-full text-xs ${
                this.getStatusColor(job.status)
            }">${job.status}</span>
                    </div>
                    <div class="grid grid-cols-2 gap-2 mt-2 text-sm">
                        <div>Channel: ${job.channel}</div>
                        <div>Progress: ${job.progress || 0}%</div>
                        <div>Started: ${new Date(job.start_time).toLocaleString()}</div>
                        <div>Duration: ${job.duration || 'N/A'}</div>
                    </div>
                </div>
            `)
            .join('');
    }

    getStatusColor(status) {
        const colors = {
            running: 'bg-blue-100 text-blue-800',
            completed: 'bg-green-100 text-green-800',
            failed: 'bg-red-100 text-red-800',
            pending: 'bg-yellow-100 text-yellow-800'
        };
        return colors[status] || 'bg-gray-100 text-gray-800';
    }

    startPolling() {
        this.updateDashboard();
        setInterval(() => this.updateDashboard(), 5000);
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    const api = new JobSchedulerAPI('');
    const dashboard = new DashboardUI(api);
});