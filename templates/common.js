const domain = "http://localhost:6789";
const filterCodeUrl = "/filter/stock/code";
const syncStockCodeUrl = "/task/stock/code";
const getCodeDataUrl = "/stock/code";
const getStockReportUrl = "/stock/report";
const updateStockReportUrl = "/stock/report";
const getPriceDataUrl = "/analyze/trend/code";
const getDataBankUrl = "/data/bank";
const getDataIndustryUrl = "/industry/bank";
const getIndustryTrendUrl = "/industry/trend"
const getIndustryBasicUrl = "/industry/basic"
const getIndustryRelationUrl = "/industry/relation"
const addSubscribeStrategyUrl = "/subscribe/strategy"
const getSubscribeStrategyUrl = "/subscribe/strategy"
const deleteSubscribeStrategyUrl = "/subscribe/strategy"
const getStockInfoUrl = "/info/stock";

const ChartPropertyMap = {
    "shareholderNumber": {
        "title": "股东人数变化趋势",
        "legend": "股东人数",
        "serieName": "股东人数"
    },
    "interestRate": {
        "title": "财报当期净息差变化趋势",
        "legend": "净息差",
        "serieName": "净息差"
    },
    "interestRatePeriod": {
        "title": "单季度净息差变化趋势",
        "legend": "净息差",
        "serieName": "净息差"
    },
    "ldRate": {
        "title": "财报当期贷款收益率&存款成本率变化趋势",
        "legend": "贷款收益率&存款成本率",
        "serieName": "贷款收益率&存款成本率"
    },
    "ldRatePeriod": {
        "title": "单季度贷款收益率&存款成本率变化趋势",
        "legend": "贷款收益率&存款成本率",
        "serieName": "贷款收益率&存款成本率"
    },
    "impairmentLoss": {
        "title": "当期信用减值损失",
        "legend": "信用减值损失",
        "serieName": "信用减值损失"
    },
    "totalBalance": {
        "title": "当期不良余额",
        "legend": "不良余额",
        "serieName": "不良余额"
    },
    "totalRate": {
        "title": "当期不良率",
        "legend": "不良率",
        "serieName": "不良率"
    },
    "newBalance": {
        "title": "当期新增不良余额",
        "legend": "新增不良",
        "serieName": "新增不良"
    },
    "newRate": {
        "title": "当期新增不良率",
        "legend": "新增不良率",
        "serieName": "新增不良率"
    },
    "coverageRate": {
        "title": "当期拨备覆盖率",
        "legend": "拨备覆盖率",
        "serieName": "拨备覆盖率"
    },
    "adequacyRate": {
        "title": "当期核心一级资本充足率",
        "legend": "核心一级资本充足率",
        "serieName": "核心一级资本充足率"
    },
    "roe": {
        "title": "当期ROE",
        "legend": "ROE",
        "serieName": "ROE"
    },
    "roa": {
        "title": "当期ROA",
        "legend": "ROA",
        "serieName": "ROA"
    },
    "rorwa": {
        "title": "当期RORWA",
        "legend": "RORWA",
        "serieName": "RORWA"
    }
}

const INDUSTRY_INFLOW = "industry_inflow";
const STOCK_INFLOW = "stock_inflow";
const MAX_OPEN_CODE_COUNT = 10;
const BTN_TYPE_INDUSTRY = "industry";
const BTN_TYPE_STOCK = "stock";

async function getPriceData(code, start_date, line_type) {
    let line_type_int = parseInt(line_type);
    const requestBody = {
        "code": code,
        "start_date": start_date,
        "k_line_type": line_type_int,
    };
    const response = await fetch(domain + getPriceDataUrl, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    });
    return response.json();
}

async function getCodeData() {
    const response = await fetch(domain + getCodeDataUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response.json();
}

async function getIndustryBasicData() {
    const response = await fetch(domain + getIndustryBasicUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response.json();
}

async function getIndustryTrendData(industry, stock, days, endDate) {
    const url = domain + getIndustryTrendUrl;
    const urlObj = new URL(url);
    urlObj.searchParams.set("days", days);
    if (industry != "") {
        urlObj.searchParams.set("industry_code", industry);
    }
    if (endDate != "") {
        urlObj.searchParams.set("end_date", endDate);
    }

    const response = await fetch(urlObj.toString(), {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response.json();
}

async function getIndustryRelationData(relationType, days) {
    const url = domain + getIndustryRelationUrl;
    const urlObj = new URL(url);
    urlObj.searchParams.set("days", days);
    if (relationType == "") {
        urlObj.searchParams.set("is_split_industry", true);
    }

    const response = await fetch(urlObj.toString(), {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response.json();
}

async function filterCode(endDate, macdFast, macdSlow, macdLength, selectedOptions, bollingPosition) {
    const url = domain + filterCodeUrl;
    let requestBody = {}
    if (endDate) {
        requestBody["date"] = endDate;
    }
    let macdFastF = parseFloat(macdFast);
    let macdSlowF = parseFloat(macdSlow);
    let macdLengthI = parseInt(macdLength);
    if ((macdFastF != 0.0 || macdSlowF!= 0.0) && macdLengthI!= 0) {
        requestBody["macd_filter"] = {}
        if (macdFastF!= 0.0) {
            requestBody["macd_filter"]["max_last_dif"] = macdFastF;
        }
        if (macdSlowF!= 0.0) {
            requestBody["macd_filter"]["max_last_dea"] = macdSlowF;
        }
        requestBody["macd_filter"]["min_length"] = macdLengthI;
    }
    if (selectedOptions.length > 0) {
        requestBody["ma_filter"] = {}
        let localMap = {}
        for (let i = 0; i < selectedOptions.length; i++) {
            let option = selectedOptions[i];
            localMap[option] = i+1;
        }
        if ("ma5" in localMap) {
            requestBody["ma_filter"]["ma5_position"] = localMap["ma5"];
        }
        if ("ma10" in localMap) {
            requestBody["ma_filter"]["ma10_position"] = localMap["ma10"];
        }
        if ("ma20" in localMap) {
            requestBody["ma_filter"]["ma20_position"] = localMap["ma20"];
        }
        if ("ma30" in localMap) {
            requestBody["ma_filter"]["ma30_position"] = localMap["ma30"];
        }
        if ("ma60" in localMap) {
            requestBody["ma_filter"]["ma60_position"] = localMap["ma60"];
        }
    }
    if (bollingPosition.length > 0) {
        requestBody["bolling_filter"] = {
            "bolling_position": bollingPosition
        }
    }
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    });
    return response.json();
}

function getSuggestOperation(s) {
    if (s == "buy") {
        return "买入";
    } else if (s == "sell") {
        return "卖出";
    }
    return "暂无";
}

function getBollingPosition(s) {
    if (s == "up") {
        return "上轨";
    } else if (s == "down") {
        return "下轨";
    } else if (s == "mid") {
        return "中轨";
    }
    return "暂无";
}

async function syncStockCode(code) {
    const url = domain + syncStockCodeUrl;
    let requestBody = {
        "code": code,
        "business_type": 2
    };
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    });
    return response;
}

async function getStockReport(code, year, reportType, disableMsg) {
    const url = domain + getStockReportUrl;
    const params = {
        "code": code,
        "year": year,
        "report_type": reportType,
        "disable_msg": disableMsg
    };

    const baseUrl = new URL(url);
    Object.entries(params).forEach(([key, value]) => {
        baseUrl.searchParams.append(key, value);
    });
    const response = await fetch(baseUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response;
}

async function updateStockReport(code, year, reportType, industryType, measurement, report, comment) {
    const url = domain + updateStockReportUrl;
    let requestBody = {
        "code": code,
        "year": parseInt(year),
        "report_type": parseInt(reportType),
        "industry_type": parseInt(industryType),
        "measurement": measurement,
        "report": report,
        "comment": comment,
    };
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    })
    return response;
}

function getReportTypeString(reportType) {
    if (reportType == 1) {
        return "一季报";
    } else if (reportType == 2) {
        return "中报";
    }   else if (reportType == 3) {
        return "三季报";
    }   else if (reportType == 4) {
        return "年报";
    }
    return "";
}

function getPreMOMReportType(reportType) {
    if (reportType == 1) {
        return 4;
    } else {
        return reportType-1;
    }
}

function getReportTypeString2(reportType) {
    if (reportType == 1) {
        return "Q1";
    } else if (reportType == 2) {
        return "Q2";
    }   else if (reportType == 3) {
        return "Q3";
    }   else if (reportType == 4) {
        return "Q4";
    }
    return "";
}

function getPeriodReportTypeString(reportType) {
    if (reportType == 1) {
        return "第一季度";
    } else if (reportType == 2) {
        return "第二季度";
    } else if (reportType == 3) {
        return "第三季度";
    } else if (reportType == 4) {
        return "第四季度";
    }
}

async function getBankTraceData(code) {
    const url = domain + getDataBankUrl;
    const params = {
        "code": code
    };

    const baseUrl = new URL(url);
    Object.entries(params).forEach(([key, value]) => {
        baseUrl.searchParams.append(key, value);
    });
    const response = await fetch(baseUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response;
}

async function getIndustryTraceData(industryType) {
    const url = domain + getDataIndustryUrl;
    const params = {
        "industry_type": industryType
    };

    const baseUrl = new URL(url);
    Object.entries(params).forEach(([key, value]) => {
        baseUrl.searchParams.append(key, value);
    });
    const response = await fetch(baseUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response;
}

async function addSubscribeStrategyData(strategy) {
    console.log(strategy);
    const url = domain + addSubscribeStrategyUrl;
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(strategy)
    });
    return response;
}

async function getSubscribeStrategyData() {
    const url = domain + getSubscribeStrategyUrl;
    const response = await fetch(url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response;
}

async function deleteSubscribeStrategyData(id) {
    const url = domain + deleteSubscribeStrategyUrl;
    const response = await fetch(url, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({"id": id})
    });
    return response;
}

async function getStockInfoData(name) {
    const url = domain + getStockInfoUrl;
    const params = {
        "name": name
    };

    const baseUrl = new URL(url);
    Object.entries(params).forEach(([key, value]) => {
        baseUrl.searchParams.append(key, value);
    });
    const response = await fetch(baseUrl, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    return response;
}



function getLoanRateTitle(name) {
    return name + "-贷款收益率";
}

function getDepositRateTitle(name) {
    return name + "-存款成本率";
}

function getUrlParamsMap() {
    const urlParams = new URLSearchParams(window.location.search);
    const paramsMap = new Map();
    urlParams.forEach((value, key) => {
        paramsMap.set(key, value);
    });
    return paramsMap;
}

function formatAmountSmart(amount) {
    if (typeof amount !== 'number' || isNaN(amount)) {
        return 'Invalid';
    }

    const absAmount = Math.abs(amount);

    // 单位配置
    const units = [
        { threshold: 1e12, unit: '万亿', divisor: 1e12 },
        { threshold: 1e8, unit: '亿', divisor: 1e8 },
        { threshold: 1e4, unit: '万', divisor: 1e4 },
        { threshold: 0, unit: '', divisor: 1 }
    ];

    // 找到合适的单位
    const selectedUnit = units.find(unit => absAmount >= unit.threshold);

    // 计算转换后的值
    const convertedValue = absAmount / selectedUnit.divisor;

    // 智能决定小数位数
    let decimalPlaces;
    if (convertedValue >= 100) {
        decimalPlaces = 0;           // 大于100，显示整数
    } else if (convertedValue >= 10) {
        decimalPlaces = 1;           // 10-100之间，显示1位小数
    } else {
        decimalPlaces = 2;           // 小于10，显示2位小数
    }

    // 格式化数字
    let formattedValue = convertedValue.toFixed(decimalPlaces);

    // 移除多余的尾随0
    if (decimalPlaces > 0) {
        formattedValue = parseFloat(formattedValue).toString();
    }

    // 添加符号和单位
    return (amount < 0 ? '-' : '') + formattedValue + selectedUnit.unit;
}