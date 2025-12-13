// ========================================
// IndiaMART Voice AI - Dashboard App
// ========================================

const API_BASE = '';  // Same origin

// ===== State =====
let sellersData = [];
let selectedSeller = null;

// ===== Sample Transcripts =====
const sampleTranscripts = {
    angry_customer: {
        text: `Agent: Good morning, thank you for calling IndiaMART support. How can I help you today?
Customer: I've had it with your service! I've been a seller for 3 years and the leads quality has been terrible for months!
Agent: I'm sorry to hear that, sir. Can I have your seller ID please?
Customer: It's 100195284. Look, I'm paying good money every month and getting garbage leads. People who aren't serious, wrong numbers, competitors calling to check prices!
Agent: I understand your frustration. Let me check your account...
Customer: This is the third time I'm calling about this! Nobody does anything! I want to cancel my subscription!
Agent: Sir, I can see your account. You're on our Star package. I see you've reported some lead issues before.
Customer: Some issues? It's a disaster! Last month I got 45 leads, only 5 were genuine. That's pathetic!
Agent: I completely understand. That conversion rate is definitely concerning. Let me escalate this to our quality team.
Customer: I don't want escalations, I want results! If this doesn't improve in 2 weeks, I'm moving to JustDial!
Agent: Sir, I'll personally ensure our quality team reviews your lead sources within 48 hours. Can I also suggest some profile optimization?
Customer: Fine, but I'm serious about leaving. This is your last chance!
Agent: I appreciate your patience, sir. You'll receive a callback from our quality team within 48 hours.`,
        sellerId: 'demo_angry_001',
        callId: 'call_' + Date.now()
    },
    happy_renewal: {
        text: `Agent: Good afternoon! Welcome to IndiaMART. How may I assist you today?
Customer: Hi! I'm calling about my subscription renewal. It's due next month.
Agent: Of course! May I have your seller ID?
Customer: Yes, it's 100263215.
Agent: Thank you! I can see your account. You've been with us for 2 years on the Leader package. How has your experience been?
Customer: Actually, really good! The leads have been consistent and I've closed several big orders through your platform.
Agent: That's wonderful to hear! Your profile shows you're in industrial machinery - a great category.
Customer: Yes, we manufacture CNC machines. The buyers coming through IndiaMART are usually serious business inquiries.
Agent: Excellent! For your renewal, I wanted to let you know we have a special offer - 15% off if you renew for 2 years.
Customer: That sounds good. What would be the total?
Agent: For the 2-year Leader package with the discount, it would be ₹85,000 per year, so ₹1,70,000 total instead of ₹2,00,000.
Customer: That's a nice saving. Let me discuss with my partner and I'll call back to confirm.
Agent: Perfect! I'll note this on your account. The offer is valid until the end of this month.
Customer: Great, thank you for the information!
Agent: You're welcome! Is there anything else I can help you with?
Customer: No, that's all. Thanks again!
Agent: Thank you for being a valued IndiaMART seller. Have a great day!`,
        sellerId: 'demo_happy_001',
        callId: 'call_' + Date.now()
    },
    technical_issue: {
        text: `Agent: Thank you for calling IndiaMART technical support. How can I help?
Customer: I can't upload my product images! It keeps showing an error.
Agent: I'm sorry for the inconvenience. Can you tell me what error message you're seeing?
Customer: It says "Upload failed - try again later" but I've been trying since yesterday!
Agent: I understand. May I have your seller ID to check your account?
Customer: It's 100579751.
Agent: Thank you. Let me check... I can see your account. Are you trying to upload from mobile or desktop?
Customer: Mobile app. The app version is 7.2.1
Agent: I see. We've had some reports of upload issues with that version. Can you try updating to the latest version 7.3.0?
Customer: I didn't know there was an update. Let me check... okay, I see it now. Updating.
Agent: Great! While that updates, make sure your images are under 5MB and in JPG or PNG format.
Customer: Oh, my images are around 8MB. Could that be the problem?
Agent: Yes, that's likely the issue! Our system has a 5MB limit per image. You can compress them or resize before uploading.
Customer: Ah okay, that makes sense. I'll resize them.
Agent: Perfect! Once you update the app and resize the images, it should work. If you still face issues, please call back.
Customer: Thanks for the help! I'll try that.
Agent: You're welcome! Is there anything else?
Customer: No, that's all. Thank you!`,
        sellerId: 'demo_tech_001',
        callId: 'call_' + Date.now()
    },
    upsell_opportunity: {
        text: `Agent: Good morning! IndiaMART customer success team. How are you today?
Customer: Good morning! I'm doing well, thanks. I had some questions about upgrading my package.
Agent: Wonderful! I'd be happy to help. May I have your seller ID?
Customer: Yes, it's 100610311.
Agent: Thank you! I can see you're currently on our Catalog package and you've been with us for 8 months.
Customer: Right. Business has been growing and I think I need more visibility. What are my options?
Agent: That's great to hear! Looking at your category - you're in packaging materials, correct?
Customer: Yes, we make corrugated boxes and packaging solutions.
Agent: Perfect. Based on your business growth, I'd recommend our Star package. It gives you 3x more lead allocation and priority listing.
Customer: What's the price difference?
Agent: Your current Catalog is ₹30,000/year. Star would be ₹60,000/year, but you get significantly more leads and better positioning.
Customer: That's double. What kind of ROI can I expect?
Agent: Based on sellers in your category, Star members typically see 2.5x more inquiries. With your average order value, that usually pays for itself within 2-3 months.
Customer: Interesting. Can I get a trial or something?
Agent: We don't have trials, but I can offer you a 3-month upgrade at a prorated rate. If you're not satisfied, you can downgrade.
Customer: That sounds fair. Let me think about it. Can you send me the details on email?
Agent: Absolutely! I'll email you the comparison and special offer details right away.
Customer: Perfect, thanks!`,
        sellerId: 'demo_upsell_001',
        callId: 'call_' + Date.now()
    }
};

// ===== Initialize =====
document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    checkConnection();
    loadSellers();
    loadDashboardData();
    loadTicketDates();
    addLog('info', 'System initialized. Ready for demo.');
});

// ===== Tab Navigation =====
function initTabs() {
    const tabs = document.querySelectorAll('.tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            // Remove active from all
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            
            // Add active to clicked
            tab.classList.add('active');
            const tabId = tab.dataset.tab;
            document.getElementById(tabId).classList.add('active');
            
            // Load data for specific tabs
            if (tabId === 'sellers') loadSellers();
            if (tabId === 'dashboard') loadDashboardData();
            if (tabId === 'tickets') loadTicketDates();
        });
    });
}

// Switch to a specific tab programmatically
function switchTab(tabId) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    
    const tab = document.querySelector(`[data-tab="${tabId}"]`);
    if (tab) {
        tab.classList.add('active');
        document.getElementById(tabId).classList.add('active');
        
        if (tabId === 'sellers') loadSellers();
        if (tabId === 'dashboard') loadDashboardData();
        if (tabId === 'tickets') loadTicketDates();
    }
}

// ===== Connection Check =====
async function checkConnection() {
    const status = document.getElementById('connectionStatus');
    try {
        const response = await fetch('/sellers');
        if (response.ok) {
            status.className = 'connection-status connected';
            status.innerHTML = '<span class="status-dot"></span><span>Connected to API</span>';
        } else {
            throw new Error('API error');
        }
    } catch (e) {
        status.className = 'connection-status error';
        status.innerHTML = '<span class="status-dot"></span><span>API Disconnected</span>';
    }
}

// ===== Logging =====
function addLog(type, message) {
    const logEntries = document.getElementById('logEntries');
    const time = new Date().toLocaleTimeString();
    const entry = document.createElement('div');
    entry.className = `log-entry ${type}`;
    entry.innerHTML = `<span class="log-time">${time}</span><span class="log-message">${message}</span>`;
    logEntries.appendChild(entry);
    logEntries.scrollTop = logEntries.scrollHeight;
}

function clearLog() {
    document.getElementById('logEntries').innerHTML = '';
    addLog('info', 'Log cleared.');
}

// ===== Sample Loader =====
function loadSample() {
    const selector = document.getElementById('sampleSelector');
    const sample = sampleTranscripts[selector.value];
    if (sample) {
        document.getElementById('transcriptInput').value = sample.text;
        document.getElementById('sellerId').value = sample.sellerId;
        document.getElementById('callId').value = sample.callId;
        addLog('info', `Loaded sample: ${selector.options[selector.selectedIndex].text}`);
    }
}

// ===== Process Transcript =====
async function processTranscript() {
    let sellerId = document.getElementById('sellerId').value.trim();
    let callId = document.getElementById('callId').value.trim() || 'call_' + Date.now();
    let transcriptInput = document.getElementById('transcriptInput').value.trim();
    let customerType = document.getElementById('customerType').value;
    let vintage = parseInt(document.getElementById('vintage').value) || 24;
    
    // Try to parse as JSON (user might paste full call JSON)
    let transcript = transcriptInput;
    try {
        const parsed = JSON.parse(transcriptInput);
        if (parsed.transcript) {
            transcript = parsed.transcript;
            addLog('info', '📋 Detected JSON format, extracted transcript field');
        }
        // Also extract other fields if available
        if (parsed.gluser_id) sellerId = parsed.gluser_id;
        if (parsed.click_to_call_id) callId = parsed.click_to_call_id;
        if (parsed.customer_type) customerType = parsed.customer_type;
        if (parsed.vintage_months) vintage = parsed.vintage_months;
        
        // Update the UI fields
        document.getElementById('sellerId').value = sellerId;
        document.getElementById('callId').value = callId;
        document.getElementById('customerType').value = customerType;
        document.getElementById('vintage').value = vintage;
    } catch (e) {
        // Not JSON, use as plain text transcript
    }
    
    if (!sellerId || !transcript) {
        addLog('error', 'Please provide Seller ID and Transcript');
        return;
    }
    
    // Update UI
    const btn = document.getElementById('processBtn');
    const indicator = document.getElementById('processingIndicator');
    btn.disabled = true;
    indicator.style.display = 'flex';
    
    addLog('info', `Starting analysis for seller: ${sellerId}`);
    
    // Prepare request body
    const requestBody = {
        gluser_id: sellerId,
        call_id: callId,
        call_text: transcript,
        customer_type: customerType,
        vintage: vintage,
        analyze: true
    };
    
    try {
        addLog('info', '📤 Sending transcript to API...');
        
        const response = await fetch('/ingest', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(requestBody)
        });
        
        if (!response.ok) {
            throw new Error(`API error: ${response.status}`);
        }
        
        const result = await response.json();
        addLog('success', '✅ Analysis complete!');
        
        // Display results
        displayAnalysisResult(result);
        
        // Refresh sellers list
        loadSellers();
        
    } catch (e) {
        addLog('error', `❌ Error: ${e.message}`);
        document.getElementById('resultsBody').innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">❌</div>
                <h3>Analysis Failed</h3>
                <p>${e.message}</p>
            </div>
        `;
    } finally {
        btn.disabled = false;
        indicator.style.display = 'none';
    }
}

// ===== Display Analysis Result =====
function displayAnalysisResult(result) {
    const resultsBody = document.getElementById('resultsBody');
    const analysis = result.analysis || {};
    
    // Extract nested fields from actual data structure
    const intent = analysis.intent || {};
    const churn = analysis.churn || {};
    const upsell = analysis.upsell || {};
    
    const sentiment = intent.sentiment || 'Unknown';
    const churnRisk = churn.is_likely_to_churn || 'unknown';
    
    const sentimentClass = `sentiment-${sentiment.toLowerCase()}`;
    const churnClass = `churn-${churnRisk.toLowerCase()}`;
    
    const issues = analysis.issues || [];
    
    // Build action items from various sources
    const actionItems = [];
    if (upsell.interested_features) {
        upsell.interested_features.forEach(f => actionItems.push(`Upsell: ${f}`));
    }
    if (churn.churn_reason) {
        actionItems.push(`Address: ${churn.churn_reason}`);
    }
    
    resultsBody.innerHTML = `
        <div class="analysis-result">
            <!-- Key Metrics -->
            <div class="result-section">
                <h3>📊 Key Metrics</h3>
                <div class="result-grid">
                    <div class="result-item">
                        <div class="result-label">Sentiment</div>
                        <div class="result-value large ${sentimentClass}">${sentiment}</div>
                    </div>
                    <div class="result-item ${churnClass}">
                        <div class="result-label">Churn Risk</div>
                        <div class="result-value large">${churnRisk}</div>
                    </div>
                    <div class="result-item">
                        <div class="result-label">Satisfaction</div>
                        <div class="result-value">${intent.satisfaction_score || 'N/A'}/5</div>
                    </div>
                    <div class="result-item">
                        <div class="result-label">Upsell Score</div>
                        <div class="result-value">${upsell.score || 'N/A'}/10</div>
                    </div>
                </div>
            </div>
            
            <!-- Summary -->
            <div class="result-section">
                <h3>📝 Summary</h3>
                <div class="result-item full">
                    <div class="result-value">${analysis.call_summary || 'No summary available'}</div>
                </div>
            </div>
            
            <!-- Churn Analysis -->
            <div class="result-section">
                <h3>📉 Churn Analysis</h3>
                <div class="result-grid">
                    <div class="result-item">
                        <div class="result-label">Renewal at Risk</div>
                        <div class="result-value">${churn.renewal_at_risk ? '⚠️ Yes' : '✅ No'}</div>
                    </div>
                    <div class="result-item">
                        <div class="result-label">Renewal Probability</div>
                        <div class="result-value">${churn.renewal_probability ? Math.round(churn.renewal_probability * 100) + '%' : 'N/A'}</div>
                    </div>
                    ${churn.churn_reason ? `
                    <div class="result-item full">
                        <div class="result-label">Churn Reason</div>
                        <div class="result-value">${churn.churn_reason}</div>
                    </div>` : ''}
                </div>
            </div>
            
            <!-- Issues -->
            <div class="result-section">
                <h3>⚠️ Issues Detected (${issues.length})</h3>
                <div class="issues-list">
                    ${issues.length > 0 ? issues.map(issue => {
                        // Handle both string and object issues
                        if (typeof issue === 'string') {
                            return `<div class="issue-item"><span class="tag issue">${issue}</span></div>`;
                        }
                        const severityClass = `severity-${(issue.severity || 'medium').toLowerCase()}`;
                        return `
                            <div class="issue-item ${severityClass}">
                                <div class="issue-header">
                                    <span class="issue-bucket">${issue.bucket || 'General'}</span>
                                    <span class="issue-severity ${severityClass}">${issue.severity || 'Medium'}</span>
                                </div>
                                <div class="issue-problem">${issue.problem || issue}</div>
                                ${issue.actionable_summary ? `<div class="issue-action">💡 ${issue.actionable_summary}</div>` : ''}
                            </div>
                        `;
                    }).join('') : '<div class="issue-item">No issues detected</div>'}
                </div>
            </div>
            
            <!-- Upsell Opportunities -->
            ${upsell.has_opportunity ? `
            <div class="result-section">
                <h3>💰 Upsell Opportunity</h3>
                <div class="result-grid">
                    <div class="result-item">
                        <div class="result-label">Score</div>
                        <div class="result-value large">${upsell.score}/10</div>
                    </div>
                    <div class="result-item">
                        <div class="result-label">Willingness to Invest</div>
                        <div class="result-value">${upsell.willingness_to_invest || 'N/A'}</div>
                    </div>
                    ${upsell.interested_features && upsell.interested_features.length > 0 ? `
                    <div class="result-item full">
                        <div class="result-label">Interested Features</div>
                        <div class="tags">
                            ${upsell.interested_features.map(f => `<span class="tag action">${f}</span>`).join('')}
                        </div>
                    </div>` : ''}
                    ${upsell.upsell_reason ? `
                    <div class="result-item full">
                        <div class="result-label">Reason</div>
                        <div class="result-value">${upsell.upsell_reason}</div>
                    </div>` : ''}
                </div>
            </div>` : ''}
            
            <!-- Agent Performance -->
            ${analysis.agent_performance ? `
            <div class="result-section">
                <h3>👨‍💼 Agent Performance</h3>
                <div class="result-item">
                    <div class="result-value large">${analysis.agent_performance}</div>
                </div>
            </div>` : ''}
            
            <!-- File Info -->
            <div class="result-section">
                <h3>🔍 File Info</h3>
                <div class="result-grid">
                    <div class="result-item">
                        <div class="result-label">Call ID</div>
                        <div class="result-value" style="font-size:12px; word-break:break-all;">${result.call_id || 'N/A'}</div>
                    </div>
                    <div class="result-item">
                        <div class="result-label">Analysis Status</div>
                        <div class="result-value">${result.analyzed ? '✅ Complete' : '⏳ Pending'}</div>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// ===== Load Sellers =====
async function loadSellers() {
    try {
        const response = await fetch('/sellers');
        if (!response.ok) throw new Error('Failed to load sellers');
        
        const data = await response.json();
        // Handle both formats: {sellers: [...]} or direct array
        sellersData = data.sellers || data;
        renderSellersList(sellersData);
        
        // Update dashboard count
        document.getElementById('totalSellers').textContent = data.total_count || sellersData.length;
        
    } catch (e) {
        document.getElementById('sellersList').innerHTML = `
            <div class="empty-state small">
                <p>Error loading sellers: ${e.message}</p>
            </div>
        `;
    }
}

function renderSellersList(sellers) {
    const list = document.getElementById('sellersList');
    
    if (!sellers || sellers.length === 0) {
        list.innerHTML = `
            <div class="empty-state small">
                <p>No sellers found. Process some transcripts first!</p>
            </div>
        `;
        return;
    }
    
    list.innerHTML = sellers.map(seller => {
        const healthLabel = seller.health_label || getHealthStatus(seller.health_score || 0).label;
        const healthClass = healthLabel.toLowerCase().replace(' ', '-');
        const issueCount = seller.open_issues || (seller.active_issues || []).length || 0;
        return `
            <div class="seller-card ${selectedSeller === seller.gluser_id ? 'selected' : ''}" 
                 onclick="selectSeller('${seller.gluser_id}')">
                <div class="seller-card-header">
                    <span class="seller-id">👤 ${seller.gluser_id}</span>
                    <span class="health-badge ${healthClass}">${healthLabel}</span>
                </div>
                <div class="seller-meta">
                    <span>📞 ${seller.total_calls || 0} calls</span>
                    <span>⚠️ ${issueCount} issues</span>
                    <span>🏷️ ${seller.customer_type || 'N/A'}</span>
                </div>
            </div>
        `;
    }).join('');
}

function getHealthStatus(score) {
    if (score >= 70) return { class: 'healthy', label: 'Healthy' };
    if (score >= 40) return { class: 'at-risk', label: 'At Risk' };
    return { class: 'critical', label: 'Critical' };
}

function filterSellers() {
    const search = document.getElementById('sellerSearch').value.toLowerCase();
    const filtered = sellersData.filter(s => 
        s.gluser_id.toLowerCase().includes(search)
    );
    renderSellersList(filtered);
}

// ===== Show Call Analysis Modal =====
async function showCallAnalysis(callId, sellerId) {
    const modal = document.getElementById('callAnalysisModal');
    const modalBody = document.getElementById('callModalBody');
    
    modal.style.display = 'flex';
    modalBody.innerHTML = `
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Loading call analysis...</p>
        </div>
    `;
    
    try {
        const response = await fetch(`/calls/${callId}`);
        if (!response.ok) throw new Error('Call analysis not found');
        
        const analysis = await response.json();
        
        // Extract nested fields
        const intent = analysis.intent || {};
        const churn = analysis.churn || {};
        const upsell = analysis.upsell || {};
        const issues = analysis.issues || [];
        
        modalBody.innerHTML = `
            <div class="call-analysis-detail">
                <!-- Call Info -->
                <div class="analysis-section">
                    <h4>📋 Call Information</h4>
                    <div class="info-grid">
                        <div class="info-item">
                            <span class="label">Call ID</span>
                            <span class="value">${analysis.call_id || callId}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">Seller ID</span>
                            <span class="value">${analysis.seller_id || sellerId}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">Analyzed At</span>
                            <span class="value">${formatDate(analysis.analyzed_at)}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">Language</span>
                            <span class="value">${analysis.original_language || 'N/A'}</span>
                        </div>
                    </div>
                </div>
                
                <!-- Key Metrics -->
                <div class="analysis-section">
                    <h4>📊 Key Metrics</h4>
                    <div class="metrics-row">
                        <div class="metric sentiment-${(intent.sentiment || 'neutral').toLowerCase()}">
                            <span class="metric-value">${intent.sentiment || 'Unknown'}</span>
                            <span class="metric-label">Sentiment</span>
                        </div>
                        <div class="metric churn-${(churn.is_likely_to_churn || 'low').toLowerCase()}">
                            <span class="metric-value">${churn.is_likely_to_churn || 'Unknown'}</span>
                            <span class="metric-label">Churn Risk</span>
                        </div>
                        <div class="metric">
                            <span class="metric-value">${intent.satisfaction_score || 'N/A'}/5</span>
                            <span class="metric-label">Satisfaction</span>
                        </div>
                        <div class="metric">
                            <span class="metric-value">${upsell.score || 0}/10</span>
                            <span class="metric-label">Upsell Score</span>
                        </div>
                    </div>
                </div>
                
                <!-- Call Summary -->
                <div class="analysis-section">
                    <h4>📝 Call Summary</h4>
                    <p class="summary-text">${analysis.call_summary || 'No summary available'}</p>
                </div>
                
                <!-- Issues Found -->
                ${issues.length > 0 ? `
                <div class="analysis-section">
                    <h4>⚠️ Issues Found (${issues.length})</h4>
                    <div class="issues-detail-list">
                        ${issues.map(issue => `
                            <div class="issue-detail-item severity-${(issue.severity || 'medium').toLowerCase()}">
                                <div class="issue-detail-header">
                                    <span class="issue-bucket-tag">${issue.bucket || 'General'}</span>
                                    <span class="issue-severity-tag ${(issue.severity || 'medium').toLowerCase()}">${issue.severity || 'Medium'}</span>
                                </div>
                                <p class="issue-problem-text">${issue.problem}</p>
                                ${issue.actionable_summary ? `<p class="issue-action-text">💡 ${issue.actionable_summary}</p>` : ''}
                            </div>
                        `).join('')}
                    </div>
                </div>
                ` : '<div class="analysis-section"><h4>✅ No Issues Found</h4><p>This call had no significant issues.</p></div>'}
                
                <!-- Churn Analysis -->
                ${churn.churn_reason ? `
                <div class="analysis-section">
                    <h4>📉 Churn Analysis</h4>
                    <div class="churn-details">
                        <p><strong>Renewal at Risk:</strong> ${churn.renewal_at_risk ? '⚠️ Yes' : '✅ No'}</p>
                        <p><strong>Renewal Probability:</strong> ${churn.renewal_probability ? Math.round(churn.renewal_probability * 100) + '%' : 'N/A'}</p>
                        <p><strong>Reason:</strong> ${churn.churn_reason}</p>
                    </div>
                </div>
                ` : ''}
                
                <!-- Upsell Opportunity -->
                ${upsell.has_opportunity ? `
                <div class="analysis-section">
                    <h4>💰 Upsell Opportunity</h4>
                    <div class="upsell-details">
                        <p><strong>Willingness to Invest:</strong> ${upsell.willingness_to_invest || 'N/A'}</p>
                        ${upsell.interested_features && upsell.interested_features.length > 0 ? `
                            <p><strong>Interested In:</strong></p>
                            <div class="tags">${upsell.interested_features.map(f => `<span class="tag action">${f}</span>`).join('')}</div>
                        ` : ''}
                        ${upsell.upsell_reason ? `<p><strong>Reason:</strong> ${upsell.upsell_reason}</p>` : ''}
                    </div>
                </div>
                ` : ''}
            </div>
        `;
        
    } catch (e) {
        modalBody.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">❌</div>
                <h3>Error Loading Analysis</h3>
                <p>${e.message}</p>
            </div>
        `;
    }
}

function closeCallModal() {
    document.getElementById('callAnalysisModal').style.display = 'none';
}

// Close modal when clicking outside
document.addEventListener('click', (e) => {
    const modal = document.getElementById('callAnalysisModal');
    if (modal && e.target === modal) {
        closeCallModal();
    }
});

// ===== Select Seller =====
async function selectSeller(sellerId) {
    selectedSeller = sellerId;
    renderSellersList(sellersData);
    
    try {
        const response = await fetch(`/sellers/${sellerId}`);
        if (!response.ok) throw new Error('Seller not found');
        
        const seller = await response.json();
        displaySellerDetail(seller);
        
    } catch (e) {
        document.getElementById('sellerDetails').innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">❌</div>
                <h3>Error</h3>
                <p>${e.message}</p>
            </div>
        `;
    }
}

function displaySellerDetail(seller) {
    const details = document.getElementById('sellerDetails');
    const healthStatus = getHealthStatus(seller.health_score || 0);
    const callHistory = seller.call_history || [];
    
    details.innerHTML = `
        <div class="seller-header">
            <div class="seller-info">
                <h3>Seller ${seller.gluser_id}</h3>
                <p class="meta">
                    ${seller.customer_type || 'Unknown Type'} • 
                    Vintage: ${seller.vintage || 0} months • 
                    Last updated: ${formatDate(seller.last_updated)}
                </p>
            </div>
            <div class="health-score">
                <div class="score ${healthStatus.class}" style="color: ${healthStatus.class === 'healthy' ? 'var(--success)' : healthStatus.class === 'at-risk' ? 'var(--warning)' : 'var(--danger)'}">
                    ${seller.health_score || 0}
                </div>
                <div class="label">Health Score</div>
            </div>
        </div>
        
        <div class="seller-stats">
            <div class="stat-box">
                <div class="value">${seller.total_calls || 0}</div>
                <div class="label">Total Calls</div>
            </div>
            <div class="stat-box">
                <div class="value">${seller.negative_interactions || 0}</div>
                <div class="label">Negative Calls</div>
            </div>
            <div class="stat-box">
                <div class="value">${(seller.active_issues || []).length}</div>
                <div class="label">Active Issues</div>
            </div>
            <div class="stat-box">
                <div class="value">${(seller.resolved_issues || []).length}</div>
                <div class="label">Resolved</div>
            </div>
        </div>
        
        <!-- Active Issues -->
        ${(seller.active_issues || []).length > 0 ? `
            <div class="result-section">
                <h3>⚠️ Active Issues</h3>
                <div class="tags">
                    ${seller.active_issues.map(issue => {
                        const issueText = typeof issue === 'string' ? issue : (issue.issue_type || issue.problem || JSON.stringify(issue));
                        return `<span class="tag issue">${issueText}</span>`;
                    }).join('')}
                </div>
            </div>
        ` : ''}
        
        <!-- Recent Calls -->
        <div class="call-history">
            <h4>📞 Recent Calls (${callHistory.length})</h4>
            ${callHistory.length > 0 ? callHistory.slice(-10).reverse().map(call => `
                <div class="call-item clickable" onclick="showCallAnalysis('${call.call_id}', '${seller.gluser_id}')">
                    <div>
                        <strong class="call-link">🔍 ${call.call_id || 'Unknown'}</strong>
                        <span style="color: var(--gray-500); margin-left: 8px;">${formatDate(call.date)}</span>
                    </div>
                    <div class="tags">
                        <span class="tag ${(call.sentiment || '').toLowerCase()}">${call.sentiment || 'Unknown'}</span>
                        <span class="tag">${call.churn_risk || 'N/A'}</span>
                    </div>
                </div>
            `).join('') : '<p>No call history available</p>'}
        </div>
        
        <!-- Call Analysis Modal -->
        <div id="callAnalysisModal" class="call-modal" style="display: none;">
            <div class="call-modal-content">
                <div class="call-modal-header">
                    <h3>📞 Call Analysis</h3>
                    <button class="close-btn" onclick="closeCallModal()">✕</button>
                </div>
                <div class="call-modal-body" id="callModalBody">
                    Loading...
                </div>
            </div>
        </div>
    `;
}

// ===== Dashboard Data =====
async function loadDashboardData() {
    console.log('Loading dashboard data...');
    
    try {
        // Load sellers data
        const sellersRes = await fetch('/sellers');
        const sellersData = await sellersRes.json();
        const sellers = sellersData.sellers || sellersData;
        const totalSellers = sellersData.total_count || sellers.length;
        const needsAttention = sellersData.needs_attention_count || sellers.filter(s => s.needs_attention).length;
        
        document.getElementById('totalSellers').textContent = totalSellers;
        document.getElementById('needsAttention').textContent = `${needsAttention} need attention`;
        
        // Render critical sellers immediately
        renderCriticalSellers(sellers.filter(s => s.health_label === 'Critical'));
        
        // Load aggregate data
        const aggRes = await fetch('/aggregates');
        const aggResponse = await aggRes.json();
        // Handle both formats: {dates: [...]} or direct array
        const aggDates = aggResponse.dates || aggResponse;
        
        if (aggDates && aggDates.length > 0) {
            const latestDate = aggDates[0];
            console.log('Loading aggregate for date:', latestDate);
            const aggDataRes = await fetch(`/aggregates/${latestDate}`);
            const aggregate = await aggDataRes.json();
            console.log('Aggregate data:', aggregate);
            
            if (aggregate) {
                // Update hero stats
                document.getElementById('totalCalls').textContent = aggregate.total_calls || 0;
                document.getElementById('totalIssues').textContent = aggregate.total_issues || 0;
                document.getElementById('avgSatisfaction').textContent = `Avg Satisfaction: ${(aggregate.avg_satisfaction_score || 0).toFixed(1)}/5`;
                document.getElementById('upsellOpps').textContent = `${aggregate.upsell_opportunities || 0} upsell opportunities`;
                document.getElementById('callsTrend').textContent = `Data from ${latestDate}`;
                
                // Render charts with correct field names
                renderIssueCategories(aggregate.feature_buckets || {});
                renderSentimentDist(aggregate.sentiment_breakdown || {});
                renderChurnRisk(aggregate.churn_risk_breakdown || {});
                renderUpsellPanel(aggregate.upsell_opportunities || 0);
                
                // Update badge counts
                const bucketCount = Object.keys(aggregate.feature_buckets || {}).length;
                document.getElementById('issueBucketCount').textContent = `${bucketCount} categories`;
            }
        } else {
            // No aggregates - show empty state
            console.log('No aggregate dates found');
            document.getElementById('totalCalls').textContent = '0';
            document.getElementById('totalIssues').textContent = '0';
            renderEmptyState('issueCategories', 'Click "Run Aggregation Now" to generate analytics');
            renderEmptyState('sentimentDist', 'Run aggregation to see sentiment data');
            renderEmptyState('churnRisk', 'Run aggregation to see churn risk');
            renderUpsellPanel(0);
        }
        
        // Load tickets count
        const ticketsRes = await fetch('/tickets');
        const ticketsResponse = await ticketsRes.json();
        // Handle both formats: {dates: [...]} or direct array
        const ticketDates = ticketsResponse.dates || ticketsResponse;
        let totalTickets = 0;
        
        if (ticketDates && ticketDates.length > 0) {
            const latestTicketDate = ticketDates[0];
            try {
                const dateTickets = await fetch(`/tickets/${latestTicketDate}`);
                const ticketsData = await dateTickets.json();
                totalTickets = ticketsData.count || (ticketsData.tickets || ticketsData || []).length;
            } catch (e) {
                console.error('Error loading tickets:', e);
            }
        }
        document.getElementById('totalTickets').textContent = totalTickets;
        
    } catch (e) {
        console.error('Dashboard load error:', e);
        document.getElementById('totalCalls').textContent = 'Error';
    }
}

function renderEmptyState(containerId, message) {
    const container = document.getElementById(containerId);
    container.innerHTML = `
        <div class="empty-state small">
            <p>${message}</p>
        </div>
    `;
}

function renderIssueCategories(buckets) {
    const container = document.getElementById('issueCategories');
    const entries = Object.entries(buckets);
    
    if (entries.length === 0) {
        container.innerHTML = '<div class="empty-state small"><p>No issue data available</p></div>';
        return;
    }
    
    // Sort by count and take top 8
    const sorted = entries.sort((a, b) => b[1].total_count - a[1].total_count).slice(0, 8);
    const max = Math.max(...sorted.map(([_, v]) => v.total_count));
    const colors = ['primary', 'warning', 'danger', 'info', 'success', 'purple', 'pink', 'teal'];
    
    container.innerHTML = `
        <div class="bar-chart">
            ${sorted.map(([label, data], i) => `
                <div class="bar-item">
                    <span class="bar-label" title="${label}">${truncate(label, 18)}</span>
                    <div class="bar-container">
                        <div class="bar-fill ${colors[i % colors.length]}" style="width: ${(data.total_count/max)*100}%"></div>
                    </div>
                    <span class="bar-value">${data.total_count} (${data.affected_sellers} sellers)</span>
                </div>
            `).join('')}
        </div>
    `;
}

function renderSentimentDist(distribution) {
    const container = document.getElementById('sentimentDist');
    const entries = Object.entries(distribution);
    
    if (entries.length === 0) {
        container.innerHTML = '<div class="empty-state small"><p>No sentiment data available</p></div>';
        return;
    }
    
    const total = entries.reduce((sum, [_, v]) => sum + v, 0);
    const sentimentColors = {
        'Positive': 'success',
        'Negative': 'danger',
        'Neutral': 'primary',
        'Mixed': 'warning'
    };
    
    // Calculate percentages for donut visualization
    const positivePercent = Math.round(((distribution['Positive'] || 0) / total) * 100);
    const negativePercent = Math.round(((distribution['Negative'] || 0) / total) * 100);
    const neutralPercent = 100 - positivePercent - negativePercent;
    
    container.innerHTML = `
        <div class="donut-chart">
            <div class="donut-visual" style="background: conic-gradient(
                var(--success) 0% ${positivePercent}%,
                var(--danger) ${positivePercent}% ${positivePercent + negativePercent}%,
                var(--gray-300) ${positivePercent + negativePercent}% 100%
            );">
                <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 60px; height: 60px; background: white; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 18px;">${total}</div>
            </div>
            <div class="donut-legend">
                ${entries.map(([label, value]) => `
                    <div class="legend-item">
                        <span class="legend-dot ${label.toLowerCase()}"></span>
                        <span>${label}: <strong>${value}</strong> (${Math.round((value/total)*100)}%)</span>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
}

function renderChurnRisk(distribution) {
    const container = document.getElementById('churnRisk');
    const entries = Object.entries(distribution);
    
    if (entries.length === 0) {
        container.innerHTML = '<div class="empty-state small"><p>No churn risk data available</p></div>';
        return;
    }
    
    const total = entries.reduce((sum, [_, v]) => sum + v, 0);
    const riskColors = {
        'high': 'danger',
        'medium': 'warning',
        'low': 'success'
    };
    
    // Order: high, medium, low
    const orderedEntries = ['high', 'medium', 'low']
        .filter(k => distribution[k] !== undefined)
        .map(k => [k, distribution[k]]);
    
    container.innerHTML = `
        <div class="bar-chart">
            ${orderedEntries.map(([label, value]) => `
                <div class="bar-item">
                    <span class="bar-label">${label.charAt(0).toUpperCase() + label.slice(1)} Risk</span>
                    <div class="bar-container">
                        <div class="bar-fill ${riskColors[label] || 'info'}" style="width: ${(value/total)*100}%"></div>
                    </div>
                    <span class="bar-value">${value} (${Math.round((value/total)*100)}%)</span>
                </div>
            `).join('')}
        </div>
        <div style="margin-top: 16px; padding: 12px; background: var(--gray-50); border-radius: 8px; font-size: 13px; color: var(--gray-600);">
            💡 <strong>${distribution['high'] || 0}</strong> sellers are at high churn risk and need immediate attention
        </div>
    `;
}

function renderCriticalSellers(sellers) {
    const container = document.getElementById('criticalSellers');
    document.getElementById('criticalCount').textContent = sellers.length;
    
    if (sellers.length === 0) {
        container.innerHTML = `
            <div class="empty-state small" style="color: var(--success);">
                <p>🎉 No critical sellers! All sellers are in good health.</p>
            </div>
        `;
        return;
    }
    
    // Sort by health score (lowest first)
    const sorted = sellers.sort((a, b) => (a.health_score || 0) - (b.health_score || 0)).slice(0, 5);
    
    container.innerHTML = `
        <div class="critical-list">
            ${sorted.map(seller => `
                <div class="critical-item" onclick="selectSeller('${seller.gluser_id}'); switchTab('sellers');" style="cursor: pointer;">
                    <div class="seller-info">
                        <span class="seller-id">👤 ${seller.gluser_id}</span>
                        <span class="seller-type">${seller.customer_type || 'Unknown'} • ${seller.open_issues || 0} issues</span>
                    </div>
                    <span class="health-badge">${seller.health_score || 0}</span>
                </div>
            `).join('')}
        </div>
        ${sellers.length > 5 ? `<p style="text-align: center; margin-top: 12px; font-size: 13px; color: var(--gray-500);">+${sellers.length - 5} more critical sellers</p>` : ''}
    `;
}

function renderUpsellPanel(opportunities) {
    const container = document.getElementById('upsellPanel');
    
    container.innerHTML = `
        <div class="upsell-stat">
            <div class="upsell-value">${opportunities}</div>
            <div class="upsell-label">Upsell Opportunities Detected</div>
            ${opportunities > 0 ? `
                <div class="upsell-hint">
                    💰 These sellers showed interest in additional features or expressed growth intentions
                </div>
            ` : `
                <div class="upsell-hint" style="background: var(--gray-50); color: var(--gray-500);">
                    Process more calls to identify upsell opportunities
                </div>
            `}
        </div>
    `;
}

// ===== Trigger Aggregation =====
async function triggerAggregation() {
    addLog('info', '📊 Triggering aggregation...');
    
    try {
        const response = await fetch('/aggregate', { method: 'POST' });
        if (!response.ok) throw new Error('Aggregation failed');
        
        addLog('success', '✅ Aggregation complete!');
        loadDashboardData();
        loadTicketDates();
        
    } catch (e) {
        addLog('error', `❌ Aggregation error: ${e.message}`);
    }
}

// ===== Tickets =====
async function loadTicketDates() {
    try {
        const response = await fetch('/tickets');
        if (!response.ok) throw new Error('Failed to load ticket dates');
        
        const data = await response.json();
        // Handle both formats: {dates: [...]} or direct array
        const dates = data.dates || data;
        const selector = document.getElementById('ticketDateFilter');
        
        if (!dates || dates.length === 0) {
            selector.innerHTML = '<option value="">No tickets yet</option>';
            return;
        }
        
        selector.innerHTML = dates.map(date => 
            `<option value="${date}">${date}</option>`
        ).join('');
        
        // Load first date's tickets
        loadTickets();
        
    } catch (e) {
        console.error('Error loading ticket dates:', e);
    }
}

async function loadTickets() {
    const date = document.getElementById('ticketDateFilter').value;
    if (!date) {
        document.getElementById('ticketsList').innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">🎫</div>
                <h3>Select a Date</h3>
                <p>Choose a date from the dropdown to view tickets</p>
            </div>
        `;
        return;
    }
    
    try {
        const response = await fetch(`/tickets/${date}`);
        if (!response.ok) throw new Error('Failed to load tickets');
        
        const data = await response.json();
        // Handle both formats: {tickets: [...]} or direct array
        const tickets = data.tickets || data;
        renderTickets(tickets);
        
    } catch (e) {
        document.getElementById('ticketsList').innerHTML = `
            <div class="empty-state">
                <p>Error: ${e.message}</p>
            </div>
        `;
    }
}

function renderTickets(tickets) {
    const list = document.getElementById('ticketsList');
    
    if (!tickets || tickets.length === 0) {
        list.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">🎫</div>
                <h3>No Tickets for This Date</h3>
            </div>
        `;
        return;
    }
    
    list.innerHTML = tickets.map(ticket => {
        // Handle priority as number or string
        const priorityNum = typeof ticket.priority === 'number' ? ticket.priority : 3;
        const priorityLabel = priorityNum <= 2 ? 'high' : priorityNum <= 4 ? 'medium' : 'low';
        const bucket = ticket.feature_bucket || ticket.category || 'Unknown';
        const affectedSellers = ticket.affected_sellers || [];
        const affectedCount = ticket.affected_count || affectedSellers.length || 0;
        
        return `
        <div class="ticket-card">
            <div class="ticket-header">
                <span class="ticket-title">🎫 ${bucket}</span>
                <span class="ticket-priority ${priorityLabel}">Priority ${priorityNum}</span>
            </div>
            <div class="ticket-body">
                <div class="ticket-meta">
                    <span>🆔 ${ticket.ticket_id || 'N/A'}</span>
                    <span>📊 ${affectedCount} issues</span>
                    <span>👥 ${affectedSellers.length} sellers</span>
                </div>
                <p class="ticket-description">${ticket.title || ticket.description?.substring(0, 200) || 'No summary'}</p>
                ${affectedSellers.length > 0 ? `
                    <div class="ticket-recommendation">
                        <strong>🎯 Affected Sellers:</strong>
                        <div class="tags" style="margin-top: 8px;">
                            ${affectedSellers.slice(0, 10).map(id => `<span class="tag" style="cursor:pointer" onclick="goToSeller('${id}')">${id}</span>`).join('')}
                            ${affectedSellers.length > 10 ? `<span class="tag">+${affectedSellers.length - 10} more</span>` : ''}
                        </div>
                    </div>
                ` : ''}
                ${ticket.top_problems && ticket.top_problems.length > 0 ? `
                    <div class="ticket-recommendation" style="margin-top: 8px;">
                        <strong>Top Issues:</strong>
                        <ul style="margin: 8px 0 0 20px; font-size: 13px;">
                            ${ticket.top_problems.slice(0, 3).map(p => `<li>${p.problem} (${p.count}x)</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}
                ${ticket.examples && ticket.examples.length > 0 ? `
                    <div class="ticket-recommendation" style="margin-top: 8px;">
                        <strong>💡 Suggested Action:</strong> ${ticket.examples[0]}
                    </div>
                ` : ''}
            </div>
        </div>
    `}).join('');
}

// ===== Reset Demo =====
function resetDemo() {
    document.getElementById('sellerId').value = 'demo_seller_001';
    document.getElementById('callId').value = '';
    document.getElementById('transcriptInput').value = '';
    document.getElementById('sampleSelector').selectedIndex = 0;
    document.getElementById('resultsBody').innerHTML = `
        <div class="empty-state">
            <div class="empty-icon">🎯</div>
            <h3>Ready to Analyze</h3>
            <p>Paste a transcript and click "Analyze Call" to see results</p>
        </div>
    `;
    clearLog();
    addLog('info', 'Demo reset. Ready for new analysis.');
}

// ===== Navigate to Seller =====
function goToSeller(sellerId) {
    // Switch to Seller Profiles tab
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    
    document.querySelector('[data-tab="sellers"]').classList.add('active');
    document.getElementById('sellers').classList.add('active');
    
    // Select the seller
    selectSeller(sellerId);
}

// ===== Utilities =====
function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    try {
        const date = new Date(dateStr);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    } catch (e) {
        return dateStr;
    }
}

function truncate(str, len) {
    if (!str) return '';
    return str.length > len ? str.substring(0, len) + '...' : str;
}
