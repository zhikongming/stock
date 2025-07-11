const domain = "http://localhost:6789";
const filterCodeUrl = "/filter/stock/code";
const syncStockCodeUrl = "/task/stock/code";
const getCodeDataUrl = "/stock/code";
const getStockReportUrl = "/stock/report";
const updateStockReportUrl = "/stock/report";

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

async function updateStockReport(code, year, reportType, industryType, measurement, report) {
    const url = domain + updateStockReportUrl;
    let requestBody = {
        "code": code,
        "year": parseInt(year),
        "report_type": parseInt(reportType),
        "industry_type": parseInt(industryType),
        "measurement": measurement,
        "report": report
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