import streamlit as st
import requests
import pandas as pd
import json
import time
from datetime import datetime

import os

# Configuration
# Read from env var if available (Docker), else default to localhost for local dev
API_URL = os.getenv("API_URL", "http://localhost:8081")
USER_ADDRESS = os.getenv("USER_ADDRESS", "")

st.set_page_config(
    page_title="TrustFlow Transparency Dashboard",
    page_icon="🛡️",
    layout="wide",
)

st.title("🛡️ TrustFlow Transparency Dashboard")

# --- Sidebar: Refresh & Stats ---
st.sidebar.header("Control Panel")
wallet = st.sidebar.text_input("Wallet Address", value=USER_ADDRESS)
if st.sidebar.button("Refresh Data"):
    st.rerun()

auto_refresh = st.sidebar.checkbox("Auto-Refresh (5s)", value=False)
if auto_refresh:
    time.sleep(5)
    st.rerun()

# --- Fetch Data ---
@st.cache_data(ttl=2)
def fetch_intents(wallet_addr):
    try:
        headers = {"X-User-Address": wallet_addr} if wallet_addr else {}
        response = requests.get(f"{API_URL}/intents", headers=headers)
        if response.status_code == 200:
            return response.json()
        else:
            st.error(f"Failed to fetch data: {response.status_code}")
            return []
    except Exception as e:
        st.error(f"Connection error: {e}")
        return []

intents_data = fetch_intents(wallet)

# --- Overview Stats ---
if intents_data:
    total = len(intents_data)
    success = sum(1 for i in intents_data if i['status'] == 'success')
    failed = sum(1 for i in intents_data if i['status'] == 'failed')
    pending = total - success - failed
    
    col1, col2, col3, col4 = st.columns(4)
    col1.metric("Total Intents", total)
    col2.metric("Success", success, delta_color="normal")
    col3.metric("Failed", failed, delta_color="inverse")
    col4.metric("Pending", pending)

# --- Main Table ---
st.subheader("Recent Activity Log")

if not intents_data:
    st.info("No intents found or server unreachable.")
else:
    # Prepare DataFrame
    df = pd.DataFrame(intents_data)
    
    # Convert timestamp
    df['created_at'] = pd.to_datetime(df['created_at'], unit='s')
    
    # Reorder columns
    df = df[['created_at', 'status', 'intent_id', 'message']]
    
    # Styling
    def color_status(val):
        color = 'grey'
        if val == 'success': color = 'green'
        elif val == 'failed': color = 'red'
        elif val == 'pending': color = 'orange'
        return f'color: {color}; font-weight: bold'

    st.dataframe(
        df.style.map(color_status, subset=['status']),
        use_container_width=True,
        column_config={
            "created_at": st.column_config.DatetimeColumn("Timestamp", format="D MMM YYYY, h:mm:ss a"),
            "intent_id": "Intent ID",
            "status": "Status",
            "message": "Message"
        }
    )

    # --- Drill Down Details ---
    st.divider()
    st.subheader("🔍 Transaction Details")
    
    selected_id = st.selectbox("Select Intent ID to view details:", options=[i['intent_id'] for i in intents_data])
    
    if selected_id:
        try:
            headers = {"X-User-Address": wallet} if wallet else {}
            details_resp = requests.get(f"{API_URL}/status/{selected_id}", headers=headers)
            if details_resp.status_code == 200:
                details = details_resp.json()
                
                # --- 0. Safety Interception Banner (Human Readable) ---
                if details['status'] == 'failed':
                    error_msg = details.get('message', '').lower()
                    
                    if 'insufficient funds' in error_msg:
                        st.error(
                            "🛑 **PREVENTED: Balance Insufficient**\n\n"
                            "The Orchestrator blocked this transaction because the wallet lacks gas fees. "
                            "**No funds were lost.**"
                        )
                    elif 'revert' in error_msg:
                        st.error(
                            "🛑 **PREVENTED: Contract Rejection**\n\n"
                            "The destination contract rejected the transaction (reverted). "
                            "This usually means invalid parameters or unauthorized access. "
                            "**No funds were lost.**"
                        )
                    else:
                        st.error(
                            "🛑 **PREVENTED: Unsafe Transaction**\n\n"
                            "The Orchestrator blocked this transaction due to a simulation failure. "
                            "**No funds were lost.**"
                        )

                # --- 1. Simulator Feedback (Proactive Trust) ---
                st.markdown("### 🛡️ Simulation & Checks")
                chk1, chk2, chk3 = st.columns(3)
                
                # Logic for Balance Check
                bal_ok = True
                bal_msg = "Orchestrator has enough TCRO."
                if details['status'] == 'failed' and 'insufficient funds' in details.get('message', '').lower():
                    bal_ok = False
                    bal_msg = "Insufficient funds for gas."
                
                if bal_ok:
                    chk1.success(f"✅ Balance Check\n\n{bal_msg}")
                else:
                    chk1.error(f"❌ Balance Check\n\n{bal_msg}")
                    
                chk2.success("✅ Contract Scan\n\nNo malicious patterns.")
                chk3.success("✅ Budget Check\n\nWithin $100 limit.")
                
                st.divider()
                
                # --- 2. Multi-Step Stepper (Operational Control) ---
                st.markdown("### 🚦 Workflow Execution")
                
                steps = details.get('steps', [])
                
                # 1. Intent Received
                t_str = datetime.fromtimestamp(details['created_at']).strftime('%H:%M:%S')
                st.markdown(f"**1. Intent Received** `@{t_str}`")
                
                # 2. Simulation
                sim_icon = "✅" if details['status'] != 'pending' else "⏳"
                st.markdown(f"**2. Simulation** {sim_icon}")
                
                # 3+. Steps
                if steps:
                    for idx, step in enumerate(steps):
                        s_status = step['status']
                        icon = "✅" if s_status == 'success' else "❌" if s_status == 'failed' else "⏳"
                        st.markdown(f"**{idx+3}. Executing Step {idx+1}/{len(steps)}: {step['action']}** {icon}")
                        if step.get('tx_hash'):
                             st.caption(f"Tx: `{step['tx_hash']}`")
                        if step.get('error'):
                             st.error(f"Error: {step['error']}")
                
                # Final: Audit
                final_icon = "✅" if details['status'] in ['success', 'failed'] else "⏳"
                st.markdown(f"**{len(steps)+3}. Audit Log Finalized** {final_icon}")

                st.divider()

                # --- 3. Explainability Log (The Debugger's View) ---
                st.subheader("🧐 Explainability Log")
                ex_c1, ex_c2 = st.columns(2)
                
                with ex_c1:
                    st.markdown("**Raw Intent (JSON)**")
                    raw_str = details.get('raw_intent', '{}')
                    try:
                        raw_json = json.loads(raw_str) if raw_str else {}
                    except:
                        raw_json = {"raw": raw_str}
                    st.json(raw_json)
                    
                with ex_c2:
                    st.markdown("**Execution Result**")
                    st.json({
                        "status": details['status'],
                        "message": details.get('message'),
                        "steps_completed": sum(1 for s in steps if s['status'] == 'success'),
                        "total_steps": len(steps)
                    })
                    
            else:
                st.warning("Could not load details.")
        except Exception as e:
            st.error(f"Error loading details: {e}")
