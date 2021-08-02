CWD=`cd $(dirname ${0}); pwd`
cd ${CWD}/../staging/datadog-agent
# github.com/DataDog/datadog-agent 0a5f72068
git diff b9f0476 -- . 
