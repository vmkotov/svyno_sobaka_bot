/* описание: получает json конфигурацию триггеров для бота */
/* автор: котов вячеслав */
/* создание: 16.01.2026 */

create or replace function svyno_sobaka_bot.get_triggers_config_json(
)
returns json
language plpgsql
as $$
declare
 v_result json;
begin
 select json_agg(
    json_strip_nulls(
        json_build_object(
            'trigger_id', t.id
            , 'trigger_name', t.name
            , 'tech_key', t.tech_key
            , 'priority', t.priority
            , 'probability', t.probability
            , 'description', t.description
            , 'patterns', coalesce(
                (select json_agg(
                    json_build_object(
                        'pattern_id', p.id
                        , 'pattern_text', p.pattern
                        , 'pattern_type', p.type
                    ) order by p.id
                ) from svyno_sobaka_bot.patterns p 
                where p.tech_key = t.tech_key and p.dt_end > now())
                , '[]'::json
            )
            , 'responses', coalesce(
                (select json_agg(
                    json_build_object(
                        'response_id', r.id
                        , 'response_text', r.response
                        , 'response_weight', r.weight
                    ) order by r.id
                ) from svyno_sobaka_bot.responses r 
                where r.tech_key = t.tech_key and r.dt_end > now())
                , '[]'::json
            )
        )
    ) order by t.priority
 ) into v_result
 from svyno_sobaka_bot.triggers t
 where t.dt_end > now();
 
 return v_result;
end;
$$;
