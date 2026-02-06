#FROM tindy2013/subconverter
FROM asdlokj1qpi23/subconverter:0.9.9

WORKDIR /base
RUN apk add tzdata --no-cache && \
  cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
  echo "Asia/Shanghai" > /etc/timezone && \
  apk del tzdata

COPY resource/groups.txt resource/rulesets.txt ./snippets/
COPY resource/pref.yml ./
COPY resource/tmpl/clash.tpl base/

EXPOSE 25500

CMD ["subconverter"]
